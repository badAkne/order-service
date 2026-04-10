package entity

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrOrderDuplicate       = errors.New("order duplicate")
	ErrProductNotFound      = errors.New("product not found")
	ErrInvalidOrderStatus   = errors.New("invalid order status")
	ErrInvalidDeliveryPrice = errors.New("invalid delivery price")
	ErrInvalidID            = errors.New("invalid id")
	ErrIncorrectParameters  = errors.New("incorrect parameters")

	ErrProductNotFoundInCatalog = errors.New("product not found in catalog")
	ErrProductPriceMismatch     = errors.New("product price mismatch")

	ErrCatalogServiceUnavailable = errors.New("catalog service unavailable")
)
