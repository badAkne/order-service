package edelivery

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/badAkne/order-service/internal/app/entity"
	ehandler "github.com/badAkne/order-service/internal/app/handler/event"
	rservice "github.com/badAkne/order-service/internal/app/service"
	"github.com/badAkne/order-service/internal/pkg/http/binding"
	butil "github.com/badAkne/order-service/pkg/broker/util"
)

type handler struct {
	serviceOrder rservice.Order
}

func NewHandler(serviceOrder rservice.Order) ehandler.Delivery {
	return &handler{serviceOrder: serviceOrder}
}

func (h *handler) CallbackDeliveryCalculated(
	ctx context.Context,
	ev *entity.EventOrderDeliveryCalculated,
	headers map[string]string,
) error {
	log.Info().
		Ctx(ctx).
		Any("msg_body", ev).
		Any("msg_headers", headers).
		Msg("Получено событие Order calculated")

	if err := binding.OnlyValidate(ev); err != nil {
		return butil.NotCriticalError(err)
	}

	guid, _ := uuid.Parse(ev.OrderID)

	err := h.serviceOrder.UpdateDeliveryPrice(ctx, guid, ev.DeliveryPrice)

	if errors.Is(err, entity.ErrNotFound) {
		log.Warn().Str("order_id", ev.OrderID).Msg("Order not found, skipping")
		return nil
	}

	if err != nil {
		return err
	}

	log.Info().
		Ctx(ctx).
		Any("msg_body", ev).
		Any("msg_headers", headers).
		Msg("Обработано событие Order calculated")

	return nil
}
