package respondent

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/badAkne/order-service/internal/pkg/http/httph"
)

type respondent struct {
	expander   Expander
	applicator Applicator
}

type HttpContext struct {
	W http.ResponseWriter
	R *http.Request
}

var ErrBadExpander = errors.New("respondent: expander is required")

func (rp *respondent) Callback(ctx any, err error) {
	manifest := rp.expander.Expand(err)
	if manifest == nil {
		return
	}

	rp.applicator.Apply(ctx, manifest)
}

func (rp *respondent) CallbackForHTTP(w http.ResponseWriter, r *http.Request, err error) {
	httpCtx := HttpContext{
		W: w,
		R: r,
	}

	rp.Callback(httpCtx, err)
}

func newRespondent(expander Expander, applicator Applicator) *respondent {
	if expander == nil {
		log.Fatal().Err(ErrBadExpander).Msg("expander is nil")
	}

	if applicator == nil {
		applicator = NewSimpleApplicator()
	}

	return &respondent{
		expander:   expander,
		applicator: applicator,
	}
}

func NewMiddleware(expander Expander, applicator Applicator) httph.Middleware {
	resp := newRespondent(expander, applicator)

	return func(c *gin.Context) {
		c.Next()

		err := httph.ErrorGet(c.Request)

		mayHandle := httph.ErrorTryAcquireHandling(c.Request)

		if err != nil && mayHandle {
			resp.CallbackForHTTP(c.Writer, c.Request, err)
		}
	}
}
