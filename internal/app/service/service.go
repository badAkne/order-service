package rservice

import (
	"context"

	"github.com/google/uuid"

	"github.com/badAkne/order-service/internal/app/entity"
)

type Order interface {
	Create(ctx context.Context, order entity.RequestOrderCreate) (entity.ResponseOrderCreate, error)
	Get(ctx context.Context, guid uuid.UUID) (entity.ResponseOrderCreate, error)
	Update(ctx context.Context, guid uuid.UUID, status string) (entity.ResponseOrderCreate, error)
	Delete(ctx context.Context, guid uuid.UUID) error
}
