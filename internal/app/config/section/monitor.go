package section

type Monitor struct {
	LogLevel    string `default:"debug" split_words:"true"`
	Environment string `default:"development"`
	Prometheus  MonitorPrometheus
	Sentry      MonitorSentry
}

type MonitorPrometheus struct {
	Enabled bool
}

type MonitorSentry struct {
	Enabled bool `default:"false"`
	DSN     string
}
