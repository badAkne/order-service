package section

import "time"

type Monitor struct {
	LogLevel      string `default:"debug" split_words:"true"`
	Environment   string `default:"development"`
	Prometheus    MonitorPrometheus
	Sentry        MonitorSentry
	OpenTelemetry MonitorOpenTelemetry `split_words:"true"`
}

type MonitorPrometheus struct {
	Enabled bool
}

type MonitorSentry struct {
	Enabled bool `default:"false"`
	DSN     string
}

type MonitorOpenTelemetry struct {
	Enabled            bool `default:"false"`
	Address            string
	MaxQueueSize       int           `default:"2048" split_words:"true"`
	MaxBatchSize       int           `default:"512"  split_words:"true"`
	SendBatchTimeout   time.Duration `default:"5s"   split_words:"true"`
	ExportTimeout      time.Duration `default:"30s"  split_words:"true"`
	SampleRatio        float64       `default:"1"    split_words:"true"`
	AddRequestHeaders  bool          `default:"true" split_words:"true"`
	AddRequestBody     bool          `default:"true" split_words:"true"`
	AddResponseHeaders bool          `default:"true" split_words:"true"`
	AddResponseBody    bool          `default:"true" split_words:"true"`
}
