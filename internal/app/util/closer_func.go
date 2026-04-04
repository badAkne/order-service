package util

import (
	"context"
	"time"
)

type CloserFunc func() error

func (f CloserFunc) Close() error {
	return f()
}

type CloserContextFunc = func(ctx context.Context) error

func NewCloserContextFunc(
	f CloserContextFunc,
	ctx context.Context,
	timeout time.Duration,
) CloserFunc {
	return func() error {
		if timeout < 0 {
			return nil
		}

		newCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return f(newCtx)
	}
}
