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

	"github.com/badAkne/order-service/internal/app/config"
	rhandler "github.com/badAkne/order-service/internal/app/handler"
	rhealth "github.com/badAkne/order-service/internal/app/handler/health"
	"github.com/badAkne/order-service/internal/app/processor"
	rprocessor "github.com/badAkne/order-service/internal/app/processor/http"
	rcpostgres "github.com/badAkne/order-service/internal/app/repository/conn/postgres"
)

type Builder struct {
	cCtx *cli.Context
	ctx  context.Context
	wg   sync.WaitGroup
	err  error
	cfg  config.Config

	connPostgres *rcpostgres.Client

	healthHandler rhandler.Health

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
}

func (b *Builder) BuildConfig(injectors ...func(c *config.Config)) {
	// TODO: Поменять на true, линтер ругается, что дается только true
	b.exec(b.ctx != nil, func(b *Builder) {
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

	log.Info().Msg("Application is initializing")
	defer log.Info().Msg("Application is completed, GoodBye!")

	for _, proc := range b.processors {
		proc.StartAsync(b.ctx, &b.wg)
	}

	b.wg.Wait()
}

func (b *Builder) BuildRepoConnPostgres() {
	b.exec(true, func(b *Builder) {
		cfg := b.cfg.Repository
		conn, err := rcpostgres.NewConn(b.ctx, cfg.Postgres)
		if err != nil {
			b.err = err
			return
		}

		b.connPostgres = conn
	})
}

func (b *Builder) BuildProcHttp() {
	b.exec(true, func(b *Builder) {
		procHttp := rprocessor.NewHTTP(b.healthHandler, nil, b.cfg.Processor.WebServer)

		b.processors = append(b.processors, procHttp)
	}, b.healthHandler)
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
