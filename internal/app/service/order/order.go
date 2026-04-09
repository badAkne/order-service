package morder

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	catalog "github.com/badAkne/order-service/internal/app/client"
	"github.com/badAkne/order-service/internal/app/entity"
	"github.com/badAkne/order-service/internal/app/repository"
	rservice "github.com/badAkne/order-service/internal/app/service"
)

type service struct {
	repoOrder     repository.Order
	catalogClient *catalog.CatalogClient
}

func NewService(repoOrder repository.Order, catalogClient *catalog.CatalogClient) rservice.Order {
	return &service{
		repoOrder:     repoOrder,
		catalogClient: catalogClient,
	}
}

func (s *service) Create(ctx context.Context, req entity.RequestOrderCreate) (entity.ResponseOrderCreate, error) {
	if err := s.validateProductsWithCatalog(ctx, req.Items); err != nil {
		return entity.ResponseOrderCreate{}, err
	}

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
	if err := s.validateStatus(status); err != nil {
		return entity.ResponseOrderCreate{}, err
	}
	existingOrder, err := s.repoOrder.Get(ctx, guid)
	if err != nil {
		return entity.ResponseOrderCreate{}, err
	}

	if existingOrder.Status == entity.OrderStatusPending && status == entity.OrderStatusPending {
		err := s.validatePricesBeforeShipping(ctx, existingOrder)
		if err != nil {
			return entity.ResponseOrderCreate{}, err
		}
	}

	var updatedOrder entity.Order
	err = s.repoOrder.OpenTx(ctx, func(txCtx context.Context) error {
		var txErr error

		_, txErr = s.repoOrder.Update(txCtx, guid, status)
		if txErr != nil {
			return txErr
		}
		updatedOrder, txErr = s.repoOrder.Get(txCtx, guid)

		return txErr
	})
	if err != nil {
		return entity.ResponseOrderCreate{}, fmt.Errorf("failed to update order: %w", err)
	}

	cartPrice := s.cartPrice(updatedOrder.Items)

	return s.convertToResponseCreate(updatedOrder, cartPrice), nil
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
			ProductGUID: item.ProductGUID,
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

func (s *service) validateProductsWithCatalog(ctx context.Context, items []entity.RequestOrderItemCreate) error {
	if items == nil {
		return errors.New("items is nil")
	}

	productsGUIDs := make([]uuid.UUID, len(items))
	for i, item := range items {
		productsGUIDs[i] = item.ProductGUID
	}

	catalogProducts, err := s.catalogClient.GetProducts(ctx, productsGUIDs)
	if err != nil {
		return entity.ErrCatalogServiceUnavailable
	}

	for _, item := range items {
		catalogProduct := catalogProducts[item.ProductGUID.String()]
		if item.UnitPrice != float32(catalogProduct.Price) {
			return entity.ErrProductPriceMismatch
		}
	}

	return nil
}

func (s *service) validateStatus(status string) error {
	switch status {
	case entity.OrderStatusCancelled:
		return nil
	case entity.OrderStatusDelivered:
		return nil
	case entity.OrderStatusPending:
		return nil
	case entity.OrderStatusShipped:
		return nil
	default:
		return entity.ErrInvalidOrderStatus
	}
}

func (s *service) validatePricesBeforeShipping(ctx context.Context, order entity.Order) error {
	productGUIDs := make([]uuid.UUID, len(order.Items))

	for i, item := range order.Items {
		productGUIDs[i] = item.ProductGUID
	}

	catalogProducts, err := s.catalogClient.GetProducts(ctx, productGUIDs)
	if err != nil {
		return fmt.Errorf("failed to check product prices: %w", err)
	}

	for _, item := range order.Items {
		catalogProduct := catalogProducts[item.ProductGUID.String()]

		if item.UnitPrice != float32(catalogProduct.Price) {
			return fmt.Errorf("cannot ship order: price for product %s has changed from %.2f to %.2f", item.ProductGUID, item.UnitPrice, catalogProduct.Price)
		}
	}

	return nil
}
