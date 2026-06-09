package order

import (
	"backend/internal/domain"
	"time"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusAccepted   OrderStatus = "accepted"
	OrderStatusRejected   OrderStatus = "rejected"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// OrderItem is a product snapshot captured at order placement.
type OrderItem struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Price     int64  `json:"price"`
	Quantity  int64  `json:"quantity"`
	Unit      string `json:"unit"`
}

type Note struct {
	AuthorID   string `json:"author_id"`
	AuthorType string `json:"author_type"`
	Message    string `json:"message"`
	CreatedAt  int64  `json:"created_at"`
}

type Order struct {
	ID              string         `json:"id"`
	WholesalerID    string         `json:"wholesaler_id"`
	DealerID        string         `json:"dealer_id"`
	PlacedByID      string         `json:"placed_by_id"`
	PlacedByType    string         `json:"placed_by_type"`
	Status          OrderStatus    `json:"status"`
	Items           []*OrderItem   `json:"items"`
	OrderValue      int64          `json:"order_value"`
	ShippingAddress *domain.Address `json:"shipping_address"`
	ETD             *time.Time     `json:"etd"`
	Notes           []*Note        `json:"notes"`
	CreatedAt       int64          `json:"created_at"`
	UpdatedAt       int64          `json:"updated_at"`
}

type OrderItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int64  `json:"quantity" binding:"required,min=1"`
}

type CreateOrderRequest struct {
	Items []*OrderItemRequest `json:"items" binding:"required,dive,required"`
	Note  string              `json:"note"`
}

type AcceptOrderRequest struct {
	ETD *time.Time `json:"etd"`
}

type AddNoteRequest struct {
	Message string `json:"message" binding:"required"`
}

type ListOrdersFilter struct {
	WholesalerID string      `json:"wholesaler_id"`
	DealerID     string      `json:"dealer_id"`
	Status       OrderStatus `json:"status"`
	PlacedByID   string      `json:"placed_by_id"`
	Page         int         `json:"page"`
	PageSize     int         `json:"page_size"`
}

type ListOrdersResponse struct {
	Orders   []*Order `json:"orders"`
	Total    int64    `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}


