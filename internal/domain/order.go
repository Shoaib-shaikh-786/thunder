package domain

import "time"

type Note struct {
	Message   string
	CreatedAt int64
}
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusAccepted   OrderStatus = "accepted"
	OrderStatusRejected   OrderStatus = "rejected"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delievered"
)

type CartType string

const (
	CartTypePrimary   CartType = "primary"
	CartTypeEphemeral CartType = "ephemeral"
)

type Order struct {
	ID                   string
	OrderStatus          OrderStatus
	Products             []*Product
	Customer             *Customer
	ShippingAddress      *Address
	BillingAddress       *Address
	CartType             CartType
	CartId               string
	OrderValue           int64
	OrderInstructions    *OrderInstructions
	ServiceabilityOption *ServiceabilityOption
	CreatedAt            int64
	UpdatedAt            int64
}

type ServiceabilityOption struct {
	Status            string
	Message           string
	CartDelieveryDate time.Time
	Items             []*ServiceabilityItem
	Shipments         []*ServiceabilityShipment
}

type ServiceabilityItem struct {
	ItemId            string
	ServiceabilityQty int32
	DelieveryDate     *time.Time
}
type ServiceabilityShipment struct {
	DelieveryDate time.Time
	DelieveryDays int32
	Items         []*Product
}
type OrderInstructions struct {
	Notes []*Note
}
