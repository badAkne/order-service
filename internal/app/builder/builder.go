package builder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	catalog "github.com/badAkne/order-service/internal/app/client"
	"github.com/badAkne/order-service/internal/app/config"
	rhandler "github.com/badAkne/order-service/internal/app/handler"
	rhealth "github.com/badAkne/order-service/internal/app/handler/health"
	rorder "github.com/badAkne/order-service/internal/app/handler/order"
	"github.com/badAkne/order-service/internal/app/processor"
	rprocessor "github.com/badAkne/order-service/internal/app/processor/http"
	mprocessor "github.com/badAkne/order-service/internal/app/processor/monitor"
	"github.com/badAkne/order-service/internal/app/repository"
	rcpostgres "github.com/badAkne/order-service/internal/app/repository/conn/postgres"
	porder "github.com/badAkne/order-service/internal/app/repository/order"
	rservice "github.com/badAkne/order-service/internal/app/service"
	morder "github.com/badAkne/order-service/internal/app/service/order"
)

type Builder struct {
	cCtx *cli.Context
	ctx  context.Context
	wg   sync.WaitGroup
	err  error
	cfg  config.Config

	connPostgres *rcpostgres.Client

	orderRepo    repository.Order
	orderSerivce rservice.Order
	orderHandler rhandler.Order

	healthHandler rhandler.Health
	catalogClient *catalog.CatalogClient

	processors []processor.Processor

	chErrors chan error
	// cancelFunc context.CancelFunc
}

func NewBuilder(cCtx *cli.Context) *Builder {
	b := Builder{
		cCtx:     cCtx,
		ctx:      context.Background(),
		chErrors: make(chan error, 4096),
	}

	ctxWithCancel, cancel := context.WithCancel(b.ctx)
	b.ctx = ctxWithCancel

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	go b.waitForSignal(sigChan, cancel)
	go b.printErrors()

	b.healthHandler = rhealth.NewHandler()
	return &b
}

func (b *Builder) exec(preCond bool, cb func(b *Builder), requiredArgs ...any) {
	if !preCond || b.err != nil {
		return
	}

	for i, reqrequiredArg := range requiredArgs {
		rv := reflect.ValueOf(reqrequiredArg)
		if !rv.IsValid() {
			b.err = fmt.Errorf("BUG: required argument #%d is nil (check dependecies)", i)
			return
		}

		if rv.Type().Kind() == reflect.Struct || !rv.IsZero() {
			continue
		}

		b.err = fmt.Errorf("BUG: required %s, but empty", rv.Type().String())
		return
	}

	cb(b)
}

func (b *Builder) buildConfig(args config.LoadArgs, injectors []func(*config.Config)) {
	args.Output = b.cCtx.App.Writer
	args.EnableSimpleLog = b.cCtx.Bool("no-json")

	config.Load(args)

	for _, injector := range injectors {
		if injector != nil {
			injector(&config.Root)
		}
	}

	b.cfg = config.Root
	mprocessor.NewSentryWriter(b.cfg.Meta.Load.Output, b.cfg.Monitor.Environment, b.cfg.Monitor.Sentry)
}

func (b *Builder) BuildConfig(injectors ...func(c *config.Config)) {
	b.exec(true, func(b *Builder) {
		b.buildConfig(config.LoadArgs{}, injectors)
	})
}

func (b *Builder) BuildConfigSimple(injectors ...func(c *config.Config)) {
	b.exec(true, func(b *Builder) {
		b.buildConfig(config.LoadArgs{SkipConfig: true}, injectors)
	})
}

func (b *Builder) Run() {
	if b.err != nil {
		log.Fatal().Err(b.err).Msg("Failed to initialize application")
	}

	defer func() {
		if b.catalogClient != nil {
			if err := b.catalogClient.Closer(); err != nil {
				log.Error().Err(err).Msg("Error closing catalog client")
			}
		}
	}()

	log.Info().Msg("Application is initializing")
	defer log.Info().Msg("Application is completed, GoodBye!")

	for _, proc := range b.processors {
		proc.StartAsync(b.ctx, &b.wg)
	}

	b.wg.Wait()
}

func (b *Builder) BuildRepoConnPostgres() {
	b.exec(b.ctx != nil, func(b *Builder) {
		cfg := b.cfg.Repository
		conn, err := rcpostgres.NewConn(b.ctx, cfg.Postgres)
		if err != nil {
			b.err = err
			return
		}

		b.connPostgres = conn
	})
}

func (b *Builder) BuildCatalogClient() {
	b.exec(true, func(b *Builder) {
		catalogAddr := b.cfg.Client.GrpcAddr

		catalogClient, err := catalog.NewCatalogClient(catalogAddr)
		if err != nil {
			b.err = fmt.Errorf("failed to create catalog client: %w", err)
			return
		}

		b.catalogClient = catalogClient
	}, b.cfg)
}

func (b *Builder) BuildRepoOrder() {
	b.exec(true, func(b *Builder) {
		repo, err := porder.NewRepo(b.ctx, b.connPostgres)
		if err != nil {
			b.err = err
			return
		}

		b.orderRepo = repo
	}, b.connPostgres)
}

func (b *Builder) BuildServiceOrder() {
	b.exec(true, func(b *Builder) {
		b.orderSerivce = morder.NewService(b.orderRepo, b.catalogClient)
	}, b.orderRepo, b.catalogClient)
}

func (b *Builder) BuildHandlerHttpOrder() {
	b.exec(true, func(b *Builder) {
		b.orderHandler = rorder.NewHandler(b.orderSerivce)
	}, b.orderSerivce)
}

func (b *Builder) BuildProcHttp() {
	b.exec(true, func(b *Builder) {
		procHttp := rprocessor.NewHTTP(b.healthHandler, b.orderHandler, nil, b.cfg.Processor.WebServer)

		b.processors = append(b.processors, procHttp)
	}, b.healthHandler, b.orderHandler)
}

func (b *Builder) BuilMonitorPrometheus() {
	b.exec(true, func(b *Builder) {
		if !b.cfg.Monitor.Prometheus.Enabled {
			log.Info().Msg("Monitoring disabled")
			return
		}

		promProc := mprocessor.NewPrometheusObserver()
		b.processors = append(b.processors, promProc)
		log.Info().Msg("Monitoring enabled")
	})
}

func (b *Builder) waitForSignal(sig chan os.Signal, cancel func()) {
	defer cancel()
	signal := <-sig
	log.Info().Msgf("Catched %s signal", signal.String())
}

func (b *Builder) printErrors() {
	for err := range b.chErrors {
		log.Error().Err(err).Msg("Catched error from errChan")
	}
}
