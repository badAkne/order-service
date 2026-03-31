package rprocessor

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/badAkne/order-service/internal/app/config/section"
	rhandler "github.com/badAkne/order-service/internal/app/handler"
)

type Processor struct {
	server *http.Server
}

func NewHTTP(
	hHealth rhandler.Health,
	// hOrder rhandler.Order,
	cfg section.ProcessorWebServer,
) *Processor {
	r := gin.Default()
	GenericRegHealthCheck(r, hHealth)
	/*
		TODO: Раскомментировать, когда будет сделан order
		v1 := router.Group("/v1")
		{
			v1GenericRegOrder(v1k hOrder)
		}
	*/

	logRoutes(r)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.ListenPort),
		Handler:      r,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}

	return &Processor{
		server: srv,
	}
}

func (p *Processor) Run() error {
	log.Println("Starting HTTP server...")
	return p.server.ListenAndServe()
}

func (p *Processor) Shutdown(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	return p.server.Shutdown(ctx)
}
