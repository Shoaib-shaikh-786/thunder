package order

import (
	"backend/internal/domain"
	"time"
)

// ── Enums ─────────────────────────────────────────────────────────────────────

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

// ── Core Models ───────────────────────────────────────────────────────────────

// OrderItem is a snapshot of a product at the time of order.
// We never reference the live product so price changes don't affect old orders.
type OrderItem struct {
	ProductID string `json:"product_id"` // reference (for linking back)
	Name      string `json:"name"`       // snapshot
	Quantity  int64  `json:"quantity"`
	Unit      string `json:"unit"`       // snapshot
	Price     int64  `json:"price"`      // snapshot in paise
}

type Note struct {
	AuthorID  string `json:"author_id"`
	AuthorType string `json:"author_type"` // dealer / salesman
	Message   string `json:"message"`
	CreatedAt int64  `json:"created_at"`
}

type Order struct {
	ID              string          `json:"id"`
	WholesalerID    string          `json:"wholesaler_id"`
	DealerID        string          `json:"dealer_id"`
	PlacedByID      string          `json:"placed_by_id"`   // dealer or salesman user id
	PlacedByType    string          `json:"placed_by_type"` // "dealer" | "salesman"
	Status          OrderStatus     `json:"status"`
	Items           []*OrderItem    `json:"items"`
	OrderValue      int64           `json:"order_value"`     // server-calculated, sum(price*qty)
	ShippingAddress *domain.Address `json:"shipping_address"` // snapshot from user at creation
	ETD             *time.Time      `json:"etd"`             // set by wholesaler on accept
	Notes           []*Note         `json:"notes"`           // dealer-added notes
	CreatedAt       int64           `json:"created_at"`
	UpdatedAt       int64           `json:"updated_at"`
}

// ── Request DTOs ──────────────────────────────────────────────────────────────

type CreateOrderRequest struct {
	Items []CreateOrderItem `json:"items" binding:"required,min=1"`
	Note  string            `json:"note"`  // optional delivery note (for salesman)
}

type CreateOrderItem struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int64  `json:"quantity"   binding:"required,min=1"`
}

type AcceptOrderRequest struct {
	ETD time.Time `json:"etd" binding:"required"` // wholesaler sets expected delivery
}

type AddNoteRequest struct {
	Message string `json:"message" binding:"required"`
}

// ── Filter / Response DTOs ────────────────────────────────────────────────────

type ListOrdersFilter struct {
	WholesalerID string
	DealerID     string      // optional — filter by specific dealer
	Status       OrderStatus // optional — filter by status
	PlacedByID   string      // optional — filter by salesman
	Page         int
	PageSize     int
}

type ListOrdersResponse struct {
	Orders   []*Order `json:"orders"`
	Total    int64    `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}