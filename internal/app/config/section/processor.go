package section

import "time"

type (
	Processor struct {
		WebServer ProcessorWebServer `split_words:"true"`
	}

	ProcessorWebServer struct {
		Host       string        `default:"localhost"`
		ListenPort uint32        `default:"9020" split_words:"true"`
		Timeout    time.Duration `default:"30s"`
	}
)
