package repository

import (
	"context"

	"github.com/badAkne/order-service/internal/app/entity"
	"github.com/google/uuid"
)

type (
	Order interface {
		Transactional

		Create(ctx context.Context, order entity.Order) (entity.Order, error)
		Get(ctx context.Context, guid uuid.UUID) (entity.Order, error)
		Update(ctx context.Context, guid uuid.UUID, status string) (entity.Order, error)
		Delete(ctx context.Context, guid uuid.UUID) error
	}

	Transactional interface {
		OpenTx(ctx context.Context, fn func(ctx context.Context) error) error
	}
)
