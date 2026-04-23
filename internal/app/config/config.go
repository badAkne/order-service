package config

import (
	"io"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/badAkne/order-service/internal/app/config/section"
)

type Config struct {
	App        section.App
	Repository section.Repository
	// Broker     section.Broker
	Processor section.Processor
	Monitor   section.Monitor
	Client    section.ClientCatalog
	Meta      Meta `ignore:"true"`
}

type Meta struct {
	WorkDir    string
	DotEnvPath string
	Load       LoadArgs
}

type LoadArgs struct {
	Output          io.Writer `json:"-"`
	EnableSimpleLog bool
	SkipConfig      bool
}

var Root Config

func Load(args LoadArgs) {
	zerolog.TimestampFieldName = "timestamp"
	zerolog.MessageFieldName = "msg"
	zerolog.TimeFieldFormat = time.RFC3339

	if args.EnableSimpleLog {
		args.Output = zerolog.ConsoleWriter{Out: args.Output}
	}

	log.Logger = createLogger(zerolog.DebugLevel, args.Output)

	log.Debug().Msg("Logger initialized with Debug level")

	if args.SkipConfig {
		log.Debug().Msg("Config loading skipped")
		return
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load condig")
	}

	Root.Meta.Load = args

	err = envconfig.Process("APP", &Root)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to process config: %s", err.Error())
	}

	level, err := zerolog.ParseLevel(Root.App.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to parse log level")
	}
	log.Logger = createLogger(level, args.Output)

	log.Info().Msg("Config and logger processed")
}

func createLogger(level zerolog.Level, output io.Writer) zerolog.Logger {
	return zerolog.New(output).
		Level(level).With().
		Timestamp().
		Logger()
}
