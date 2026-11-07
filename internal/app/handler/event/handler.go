package ehandler

import (
	"context"

	"github.com/badAkne/order-service/internal/app/entity"
)

type (
	Delivery interface {
		CallbackDeliveryCalculated(ctx context.Context, ev *entity.EventOrderDeliveryCalculated, headers map[string]string) error
	}
)
