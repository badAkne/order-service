package rservice

import (
	"context"

	"github.com/badAkne/order-service/internal/app/entity"
	"github.com/google/uuid"
)

type Order interface {
	Create(ctx context.Context, order entity.RequestOrderCreate) (entity.ResponseOrderCreate, error)
	Get(ctx context.Context, guid uuid.UUID) (entity.ResponseOrderCreate, error)
	Update(ctx context.Context, guid uuid.UUID, status string) (entity.ResponseOrderCreate, error)
	Delete(ctx context.Context, guid uuid.UUID) error
}
