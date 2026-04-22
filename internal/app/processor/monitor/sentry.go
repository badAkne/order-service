package mprocessor

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/badAkne/order-service/internal/app/config/section"
	"github.com/badAkne/order-service/internal/app/processor"
)

type sentryProc struct {
	w   io.Writer
	cfg section.MonitorSentry
}

type sentryProcWriter sentryProc

func NewSentryWriter(
	defaultWriter io.Writer,
	env string,
	cfg section.MonitorSentry,
) processor.Processor {
	p := sentryProc{
		w:   defaultWriter,
		cfg: cfg,
	}

	if !cfg.Enabled {
		return &p
	}

	p.cfg.Enabled = false

	if cfg.DSN == "" {
		log.Error().Msg("dsn for sentry is nil")
		return &p
	}

	var w io.Writer = (*sentryProcWriter)(&p)

	w = zerolog.MultiLevelWriter(defaultWriter, w)

	serviceName := "catalog-service"
	if env != "" {
		serviceName += "-" + strings.ToLower(env)
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:        cfg.DSN,
		ServerName: serviceName,
	})
	if err != nil {
		log.Error().Err(err).Msg("unable to init sentry")
		return &p
	}

	log.Logger = log.Output(w)

	p.cfg.Enabled = true

	return &p
}

func (p *sentryProc) StartAsync(_ context.Context, _ *sync.WaitGroup) {
	if !p.cfg.Enabled {
		return
	}

	log.Info().Msg("Sentry logger hook has been initialized and started")
}

func (w *sentryProcWriter) Write(p []byte) (int, error) {
	return w.WriteLevel(log.Logger.GetLevel(), p)
}

func (w *sentryProcWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	switch level {
	case zerolog.ErrorLevel:
	case zerolog.FatalLevel:
	case zerolog.PanicLevel:
	default:
		return len(p), nil
	}

	e := sentry.Event{
		Timestamp: time.Now(),
		Logger:    "zerolog",
		Extra:     make(map[string]any),
	}

	err := json.Unmarshal(p, &e.Extra)
	if err != nil {
		log.Warn().Err(err).Msg("unable to unmarshal json")
		return len(p), err
	}

	for k, v := range e.Extra {
		switch k {
		case "message":
			if m, ok := v.(string); ok {
				e.Message = m
			}
		case "error":
			if err, ok := v.(error); ok {
				e.SetException(err, -1)
			}
		}
	}

	sentry.CaptureEvent(&e)

	if level == zerolog.FatalLevel || level == zerolog.PanicLevel {
		sentry.Flush(5 * time.Second)
	}

	return len(p), nil
}
