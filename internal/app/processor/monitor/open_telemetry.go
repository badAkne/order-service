package mprocessor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/badAkne/order-service/internal/app/config/section"
	"github.com/badAkne/order-service/internal/app/pkg/constant"
	"github.com/badAkne/order-service/internal/app/processor"
	"github.com/badAkne/order-service/internal/app/util"
)

type (
	openTelemetryProc struct {
		traceProvider *trace.TracerProvider
	}

	openTelemetryErrorHandler struct{}
)

func NewOpenTelemetryController(
	ctx context.Context,
	env string,
	cfg section.MonitorOpenTelemetry,
) processor.Processor {
	var p openTelemetryProc

	serviceName := constant.AppName

	if env != "" {
		serviceName += "-" + strings.ToLower(env)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(serviceName)))
	if err != nil {
		cleanupAndFatal(cancel, err)
		return nil
	}

	//nolint:staticcheck
	conn, err := grpc.DialContext(ctx, cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		cleanupAndFatal(cancel, err)
		return nil
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		cleanupAndFatal(cancel, fmt.Errorf("failed to create exporter: %w", err))
		return nil
	}
	if cfg.SampleRatio < 0 {
		cfg.SampleRatio = 0
	} else if cfg.SampleRatio > 1 {
		cfg.SampleRatio = 1
	}

	bsp := trace.NewBatchSpanProcessor(
		exporter,
		trace.WithExportTimeout(cfg.ExportTimeout),
		trace.WithBatchTimeout(cfg.SendBatchTimeout),
		trace.WithMaxExportBatchSize(cfg.MaxBatchSize),
		trace.WithMaxQueueSize(cfg.MaxQueueSize),
	)

	p.traceProvider = trace.NewTracerProvider(
		trace.WithSampler(trace.TraceIDRatioBased(cfg.SampleRatio)),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(p.traceProvider)

	otel.SetTextMapPropagator(jaeger.Jaeger{})

	otel.SetErrorHandler(openTelemetryErrorHandler{})

	return &p
}

func (p *openTelemetryProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	processor.WatchForShutdown(ctx, wg, util.CloserFunc(p.shutdown))
}

func (openTelemetryErrorHandler) Handle(err error) {
	log.Error().Err(err).Msg("OpenTelemetry error")
}

func (p *openTelemetryProc) shutdown() error {
	const errMsg = "Failed to shutdown trace provider"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := p.traceProvider.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg(errMsg)
		return err
	}
	return nil
}

func cleanupAndFatal(cancel func(), err error) {
	cancel()
	log.Fatal().Err(err).Msg("Failed to initialize OpenTelemetry")
}
