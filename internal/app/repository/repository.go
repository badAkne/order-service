package repository

import "context"

type (
	Transactional interface {
		OpenTx(ctx context.Context, fn func(ctx context.Context) error) error
	}
)
