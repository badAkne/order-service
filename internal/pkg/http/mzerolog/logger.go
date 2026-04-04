package mzerolog

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/badAkne/order-service/internal/pkg/http/httph"
)

type CallbackExtractorString = func(r *http.Request) string

type CallbackExtractorAny = func(r *http.Request) any

type extractorStr struct {
	key string
	ext CallbackExtractorString
}

type extractorAny struct {
	key string
	ext CallbackExtractorAny
}

type middleware struct {
	log zerolog.Logger

	fromOptions struct {
		skipper func(r *http.Request) bool

		extStrOnSuccess []extractorStr
		extAnyOnSuccess []extractorAny

		extStrOnFail []extractorStr
		extAnyOnFail []extractorAny
	}
}

func (m *middleware) Callback(c *gin.Context) {
	const (
		TailSuccess = " finished without error"
		TailFail    = " finished (or aborted) with error"
	)

	start := time.Now()

	c.Next()

	r := c.Request

	err := httph.ErrorGet(r)

	execTime := time.Since(start)

	if m.fromOptions.skipper(r) {
		return
	}

	var mb strings.Builder
	mb.Grow(48 + len(r.RequestURI))
	mb.WriteString(r.Method)
	mb.WriteByte(' ')
	mb.WriteString(r.RequestURI)

	var ev *zerolog.Event
	var extString []extractorStr
	var extAny []extractorAny

	if err == nil {
		mb.WriteString(TailSuccess)
		ev = m.log.Debug()
		extString = m.fromOptions.extStrOnSuccess
		extAny = m.fromOptions.extAnyOnSuccess
	} else {
		mb.WriteString(TailFail)
		ev = m.log.Error()
		extString = m.fromOptions.extStrOnFail
		extAny = m.fromOptions.extAnyOnFail
	}

	m.applyExtractors(r, ev, extString, extAny)

	ev.Err(err)
	ev.Str("exec_time", execTime.String())
	ev.Str("client_ip", r.RemoteAddr)

	ev.Msg(mb.String())
}

func NewMiddleware(opts ...Option) gin.HandlerFunc {
	m := middleware{
		log: log.Logger,
	}

	m.fromOptions.skipper = defaultSkipper

	for _, opt := range opts {
		opt(&m)
	}

	return m.Callback
}

func defaultSkipper(_ *http.Request) bool {
	return false
}

func (m *middleware) applyExtractors(
	r *http.Request,
	ev *zerolog.Event,
	extractorsStr []extractorStr,
	extractorsAny []extractorAny,
) {
	for _, extStr := range extractorsStr {
		key := extStr.key
		valueStr := extStr.ext(r)
		if valueStr != "" {
			ev.Str(key, valueStr)
		}
	}

	for _, extAny := range extractorsAny {
		key := extAny.key
		valueAny := extAny.ext(r)
		if valueAny != nil {
			ev.Any(key, valueAny)
		}
	}
}
