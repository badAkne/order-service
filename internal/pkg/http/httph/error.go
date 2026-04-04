package httph

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type _ContextKeyError struct{}

type _ContextValueError struct {
	err       error
	detail    string
	isHandled bool
}

func errorPrepare(ctx context.Context) context.Context {
	errCtx := new(_ContextValueError)
	return context.WithValue(ctx, _ContextKeyError{}, errCtx)
}

func errorGet(ctx context.Context) error {
	errV, _ := ctx.Value(_ContextKeyError{}).(*_ContextValueError)

	if errV != nil {
		return errV.err
	}

	return nil
}

func errorGetDetail(ctx context.Context) string {
	errD, _ := ctx.Value(_ContextKeyError{}).(*_ContextValueError)

	if errD != nil {
		return errD.detail
	}

	return ""
}

func errorApply(ctx context.Context, err error) {
	errV, _ := ctx.Value(_ContextKeyError{}).(*_ContextValueError)

	if errV != nil {
		errV.err = err
	}
}

func errorApplyDetail(ctx context.Context, detail string) {
	errD, _ := ctx.Value(_ContextKeyError{}).(*_ContextValueError)

	if errD != nil {
		errD.detail = detail
	}
}

func errorTryAcquireHandling(ctx context.Context) bool {
	errV, _ := ctx.Value(_ContextKeyError{}).(*_ContextValueError)

	if errV != nil && errV.isHandled {
		return false
	}

	errV.isHandled = true
	return true
}

func ErrorPrepare(r *http.Request) *http.Request {
	return r.WithContext(errorPrepare(r.Context()))
}

func ErrorGet(r *http.Request) error {
	ctx := r.Context()

	return errorGet(ctx)
}

func ErrorGetDetail(r *http.Request) string {
	ctx := r.Context()

	return errorGetDetail(ctx)
}

func ErrorApply(r *http.Request, err error) {
	ctx := r.Context()

	errorApply(ctx, err)
}

func ErrorApplyDetail(r *http.Request, detail string) {
	ctx := r.Context()
	errorApplyDetail(ctx, detail)
}

func ErrorTryAcquireHandling(r *http.Request) bool {
	ctx := r.Context()

	return errorTryAcquireHandling(ctx)
}

type Middleware = gin.HandlerFunc

func NewErrorMiddleware() Middleware {
	return func(ctx *gin.Context) {
		ctx.Request = ErrorPrepare(ctx.Request)

		ctx.Next()
	}
}
