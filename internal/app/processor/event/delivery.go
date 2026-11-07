package eprocessor

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/badAkne/order-service/internal/app/entity"
	ehandler "github.com/badAkne/order-service/internal/app/handler/event"
	"github.com/badAkne/order-service/internal/app/processor"
	"github.com/badAkne/order-service/pkg/broker"
)

type deliveryCalculatedProc struct {
	h   ehandler.Delivery
	bus broker.Bus[entity.EventOrderDeliveryCalculated]
}

func NewDeliveryCalculatedEventsCatcher(
	h ehandler.Delivery,
	bus broker.Bus[entity.EventOrderDeliveryCalculated],
) processor.Processor {
	return &deliveryCalculatedProc{h, bus}
}

func (p deliveryCalculatedProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	const S1 = "Не удалось подписаться на топик"
	const S2 = "Подписка на топик инициализирована"

	err := p.bus.Subscribe(ctx, wg, p.h.CallbackDeliveryCalculated)
	if err != nil {
		log.Info().Err(err).Str("topic_name", "order.delivery.calculated").Msg(S1)
		return
	}

	log.Info().Str("topic_name", "order.delivery.calculated").Msg(S2)
}
