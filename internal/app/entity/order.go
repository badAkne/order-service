package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	OrderStatusPending   = "pending"
	OrderStatusCancelled = "cancelled"
	OrderStatusDelivered = "delivered"
	OrderStatusShipped   = "shipped"

	EventOrderHeaderTypeKey = "type"

	EventOrderHeaderCreated = "order.created"
)

type Order struct {
	ID            uint      `gorm:"column:id;not null;unique;autoIncrement"`
	GUID          uuid.UUID `gorm:"column:guid;not null;primaryKey"`
	UserGUID      uuid.UUID `gorm:"column:user_guid;type:uuid;default:null"`
	TotalPrice    float32   `gorm:"column:total_price;type:decimal(12,3);default:0"`
	DeliveryPrice float32   `gorm:"column:delivery_price;type:decimal(12,3);default:0"`
	Currency      string    `gorm:"column:currency;type:text;not null"`
	Status        string    `gorm:"column:status;type:text;not null"`
	CreatedAt     time.Time `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null"`

	Items []OrderItem `gorm:"foreignKey:OrderGUID;constraint:OnDelete:CASCADE"`
}

func (Order) TableName() string {
	return "orders"
}

type OrderItem struct {
	ID          int64     `gorm:"column:id;type:bigint;not null;unique;autoIncrement"`
	GUID        uuid.UUID `gorm:"column:guid;type:uuid;not null;primaryKey"`
	OrderGUID   uuid.UUID `gorm:"column:order_guid;type:uuid;not null"`
	ProductGUID uuid.UUID `gorm:"column:product_guid;type:uuid;not null"`
	Quantity    int64     `gorm:"column:quantity;type:int;not null"`
	UnitPrice   float32   `gorm:"column:unit_price;type:decimal(12,3);not null"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

type RequestOrderCreate struct {
	UserGUID      uuid.UUID `json:"user_guid" binding:"required,uuid"`
	DeliveryPrice float32   `json:"delivery_price" binding:"required,gt=0"`
	Currency      string    `json:"currency" binding:"required,iso4217"`
	Items         []RequestOrderItemCreate
}

type RequestOrderItemCreate struct {
	ProductGUID uuid.UUID `json:"product_guid" binding:"required,uuid"`
	Quantity    int       `json:"quantity" binding:"required,gt=0"`
	UnitPrice   float32   `json:"unit_price" binding:"required,gt=0"`
}

type RequestUpdateOrder struct {
	Status string `json:"status" binding:"required"`
}

type ResponseOrderCreate struct {
	GUID          uuid.UUID `json:"order_guid"`
	UserGUID      uuid.UUID `json:"user_guid"`
	CartPrice     float32   `json:"cart_price"`
	TotalPrice    float32   `json:"total_price"`
	DeliveryPrice float32   `json:"delivery_price"`
	Currency      string    `jsom:"currency"`
	Status        string    `json:"status"`
	CreateAt      string    `json:"created_at"`

	Items []OrderItem `json:"items"`
}

type ResponseOrderItemCreate struct {
	GUID        uuid.UUID `json:"order_item_guid"`
	OrderGUID   uuid.UUID `json:"order_guid"`
	ProductGUID uuid.UUID `json:"product_guid"`
	Quantity    int64     `json:"quantity"`
	UnitPrice   float32   `json:"unit_price"`
}

type EventOrderCreated struct {
	OrderID     string           `json:"order_id"`
	UserID      string           `json:"user_id"`
	TotalAmount float64          `json:"total_amount"`
	Currency    string           `json:"currency"`
	Items       []EventOrderItem `json:"items"`
	CreatedAt   string           `json:"created_at"`
}

type EventOrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
