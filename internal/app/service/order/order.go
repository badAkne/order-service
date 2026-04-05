package morder

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/badAkne/order-service/internal/app/entity"
	"github.com/badAkne/order-service/internal/app/repository"
	rservice "github.com/badAkne/order-service/internal/app/service"
)

type service struct {
	repoOrder repository.Order
}

func NewService(repoOrder repository.Order) rservice.Order {
	return &service{
		repoOrder: repoOrder,
	}
}

func (s *service) Create(ctx context.Context, req entity.RequestOrderCreate) (entity.ResponseOrderCreate, error) {
	var cartPrice float32
	now := time.Now()

	order := entity.Order{
		GUID:          uuid.New(),
		UserGUID:      req.UserGUID,
		DeliveryPrice: req.DeliveryPrice,
		Currency:      req.Currency,
		Status:        "pending",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	order.Items, cartPrice = s.convertReqProductItemsToProductItems(req.Items, order.GUID)
	order.TotalPrice = cartPrice + order.DeliveryPrice

	var createdOrder entity.Order

	err := s.repoOrder.OpenTx(ctx, func(txCtx context.Context) error {
		var txErr error
		createdOrder, txErr = s.repoOrder.Create(txCtx, order)
		return txErr
	})
	if err != nil {
		return entity.ResponseOrderCreate{}, fmt.Errorf("unable to create order: %w", err)
	}

	fmt.Println(createdOrder)

	return s.convertToResponseCreate(createdOrder, cartPrice), nil
}

func (s *service) Get(ctx context.Context, guid uuid.UUID) (entity.ResponseOrderCreate, error) {
	var order entity.Order

	err := s.repoOrder.OpenTx(ctx, func(txCtx context.Context) error {
		var txErr error
		order, txErr = s.repoOrder.Get(txCtx, guid)
		return txErr
	})
	if err != nil {
		return entity.ResponseOrderCreate{}, fmt.Errorf("unable to get order: %w", err)
	}

	cartPrice := s.cartPrice(order.Items)

	return s.convertToResponseCreate(order, cartPrice), nil
}

func (s *service) Update(ctx context.Context, guid uuid.UUID, status string) (entity.ResponseOrderCreate, error) {
	var order entity.Order

	err := s.repoOrder.OpenTx(ctx, func(txCtx context.Context) error {
		var txErr error

		_, txErr = s.repoOrder.Update(txCtx, guid, status)
		if txErr != nil {
			return txErr
		}
		order, txErr = s.repoOrder.Get(txCtx, guid)

		return txErr
	})
	if err != nil {
		return entity.ResponseOrderCreate{}, fmt.Errorf("failed to update order: %w", err)
	}

	cartPrice := s.cartPrice(order.Items)

	return s.convertToResponseCreate(order, cartPrice), nil
}

func (s *service) Delete(ctx context.Context, guid uuid.UUID) error {
	err := s.repoOrder.OpenTx(ctx, func(txCtx context.Context) error {
		return s.repoOrder.Delete(txCtx, guid)
	})
	if err != nil {
		return fmt.Errorf("unable to delete order: %w", err)
	}

	return nil
}

func (s *service) convertReqProductItemsToProductItems(items []entity.RequestOrderItemCreate, orderGUID uuid.UUID) ([]entity.OrderItem, float32) {
	productItems := make([]entity.OrderItem, 0, len(items))
	var cartPrice float32

	for _, item := range items {
		productItems = append(productItems, entity.OrderItem{
			GUID:        uuid.New(),
			OrderGUID:   orderGUID,
			ProductGUID: orderGUID,
			Quantity:    int64(item.Quantity),
			UnitPrice:   item.UnitPrice,
		})

		cartPrice += item.UnitPrice * float32(item.Quantity)
	}

	return productItems, cartPrice
}

func (s *service) convertToResponseCreate(order entity.Order, cartPrice float32) entity.ResponseOrderCreate {
	response := entity.ResponseOrderCreate{
		GUID:          order.GUID,
		UserGUID:      order.UserGUID,
		Status:        order.Status,
		DeliveryPrice: order.DeliveryPrice,
		Currency:      order.Currency,
		CreateAt:      order.CreatedAt.Format(time.RFC3339),
		Items:         make([]entity.OrderItem, 0, len(order.Items)),
		CartPrice:     cartPrice,
		TotalPrice:    cartPrice + order.DeliveryPrice,
	}

	for _, item := range order.Items {
		response.Items = append(response.Items, entity.OrderItem{
			GUID:        item.GUID,
			OrderGUID:   item.OrderGUID,
			ID:          item.ID,
			ProductGUID: item.ProductGUID,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}

	return response
}

func (s *service) cartPrice(items []entity.OrderItem) float32 {
	var cartPrice float32

	for _, item := range items {
		cartPrice += item.UnitPrice * item.UnitPrice
	}

	return cartPrice
}
