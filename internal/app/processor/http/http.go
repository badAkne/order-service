package rprocessor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/badAkne/order-service/internal/app/config/section"
	rhandler "github.com/badAkne/order-service/internal/app/handler"
	"github.com/badAkne/order-service/internal/app/pkg/constant"
	"github.com/badAkne/order-service/internal/app/processor"
	"github.com/badAkne/order-service/internal/app/util"
	"github.com/badAkne/order-service/internal/pkg/http/httph"
	"github.com/badAkne/order-service/internal/pkg/http/mzerolog"
)

type Processor struct {
	server *http.Server
	addr   string
}

func NewHTTP(
	hHealth rhandler.Health,
	hOrder rhandler.Order,
	_ []httph.Middleware,
	cfg section.ProcessorWebServer,
) *Processor {
	r := gin.New()

	r.Use(
		otelgin.Middleware(constant.AppName,
			otelgin.WithFilter(func(r *http.Request) bool {
				return !util.IsFilteredWithHttp(r)
			})),

		httph.NewErrorMiddleware(),

		mzerolog.NewMiddleware(
			mzerolog.WithSkipper(util.IsFilteredWithHttp),
		),

		makeErrorMiddleware(),
	)
	GenericRegHealthCheck(r, hHealth)
	GenericRegPprof(r)
	GenericRegMetrics(r)

	v1 := r.Group("/v1")
	{
		v1GenericRegOrder(v1, hOrder)
	}

	logRoutes(r)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.ListenPort),
		Handler:      r,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}

	return &Processor{
		server: srv,
		addr:   fmt.Sprintf("%s:%d", cfg.Host, cfg.ListenPort),
	}
}

func (p *Processor) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "tcp", p.addr)
	if err != nil {
		log.Fatal().Err(err).Str("listen_addr", p.addr).Msg("Failed to start listening TCP addr for HTTP servver")
		return
	}

	log.Info().Str("listen_addr", p.addr).Msg("Listening of TCP addr for HTTP server has been started")

	go p.serve(l)

	processor.WatchForShutdown(ctx, wg, util.CloserFunc(l.Close))

	processor.WatchForShutdown(ctx, wg, util.NewCloserContextFunc(p.server.Shutdown, context.Background(), 5*time.Second))
}

func (h *Processor) serve(l net.Listener) {
	_ = h.server.Serve(l)
}
