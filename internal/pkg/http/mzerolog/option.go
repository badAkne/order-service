package mzerolog

import (
	"net/http"

	"github.com/rs/zerolog"
)

type Option = func(m *middleware)

func WithLogger(l zerolog.Logger) Option {
	return func(m *middleware) {
		m.log = l
	}
}

func WithSkipper(skipper func(r *http.Request) bool) Option {
	return func(m *middleware) {
		if skipper != nil {
			m.fromOptions.skipper = skipper
		}
	}
}

func WithStringExtractor(key string, callback CallbackExtractorString) Option {
	return func(m *middleware) {
		if key != "" && callback != nil {
			ext := extractorStr{key, callback}
			m.fromOptions.extStrOnSuccess = append(m.fromOptions.extStrOnSuccess, ext)
			m.fromOptions.extStrOnFail = append(m.fromOptions.extStrOnFail, ext)
		}
	}
}

func WithStringExtractorOnSuccess(key string, callback CallbackExtractorString) Option {
	return func(m *middleware) {
		if key != "" && callback != nil {
			ext := extractorStr{key, callback}
			m.fromOptions.extStrOnSuccess = append(m.fromOptions.extStrOnSuccess, ext)
		}
	}
}

func WithStringExtractorOnFail(key string, callback CallbackExtractorString) Option {
	return func(m *middleware) {
		if key != "" && callback != nil {
			ext := extractorStr{key, callback}
			m.fromOptions.extStrOnFail = append(m.fromOptions.extStrOnFail, ext)
		}
	}
}

func WithAnyExtractor(key string, callback CallbackExtractorAny) Option {
	return func(m *middleware) {
		if key != "" && callback != nil {
			ext := extractorAny{key, callback}
			m.fromOptions.extAnyOnSuccess = append(m.fromOptions.extAnyOnSuccess, ext)
			m.fromOptions.extAnyOnFail = append(m.fromOptions.extAnyOnFail, ext)
		}
	}
}

func WithAnyExtractorOnSuccess(key string, callback CallbackExtractorAny) Option {
	return func(m *middleware) {
		if key != "" && callback != nil {
			ext := extractorAny{key, callback}
			m.fromOptions.extAnyOnSuccess = append(m.fromOptions.extAnyOnSuccess, ext)
		}
	}
}

func WithAnyExtractorOnFail(key string, callback CallbackExtractorAny) Option {
	return func(m *middleware) {
		if key != "" && callback != nil {
			ext := extractorAny{key, callback}
			m.fromOptions.extAnyOnFail = append(m.fromOptions.extAnyOnFail, ext)
		}
	}
}
