package order

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ── Create ────────────────────────────────────────────────────────────────────

// CreateOrder is called by buyer or field agent.
// placedByID   = the user placing the order
// placedByType = "buyer" | "field_agent"
func (s *Service) CreateOrder(
	ctx context.Context,
	tenantID, placedByID, placedByType string,
	req CreateOrderRequest,
) (*Order, error) {
	// Validate the user placing the order has a saved address — required for shipping
	addr, err := s.repo.GetUserAddress(ctx, placedByID)
	if err != nil {
		return nil, fmt.Errorf("fetch user address: %w", err)
	}
	if addr == nil {
		return nil, fmt.Errorf("user has no saved address; please update profile before placing an order")
	}

	// Snapshot each product from the tenant's catalog
	var items []*OrderItem
	var orderValue int64

	for _, reqItem := range req.Items {
		snapshot, err := s.repo.GetProductSnapshot(ctx, reqItem.ProductID, tenantID)
		if err != nil {
			return nil, err
		}
		snapshot.Quantity = reqItem.Quantity
		orderValue += snapshot.Price * reqItem.Quantity
		items = append(items, snapshot)
	}

	now := time.Now().UnixMilli()
	o := &Order{
		ID:              uuid.New().String(),
		TenantID:        tenantID,
		BuyerID:         placedByID,
		PlacedByID:      placedByID,
		PlacedByType:    placedByType,
		Status:          OrderStatusPending,
		Items:           items,
		OrderValue:      orderValue,
		ShippingAddress: addr, // snapshot at order time
		Notes:           []*Note{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Attach delivery note if provided
	if req.Note != "" {
		o.Notes = append(o.Notes, &Note{
			AuthorID:   placedByID,
			AuthorType: placedByType,
			Message:    req.Note,
			CreatedAt:  now,
		})
	}

	if err := s.repo.Create(ctx, o); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return o, nil
}

// ── Status Transitions ────────────────────────────────────────────────────────

// AcceptOrder — Admin only. Sets ETD at this point.
func (s *Service) AcceptOrder(ctx context.Context, tenantID, orderID string, req AcceptOrderRequest) error {
	o, err := s.getAndValidate(ctx, orderID, tenantID)
	if err != nil {
		return err
	}
	if o.Status != OrderStatusPending {
		return fmt.Errorf("can only accept a pending order, current status: %s", o.Status)
	}
	return s.repo.UpdateStatus(ctx, orderID, tenantID, OrderStatusAccepted, req.ETD)
}

// RejectOrder — Admin only.
func (s *Service) RejectOrder(ctx context.Context, tenantID, orderID string) error {
	o, err := s.getAndValidate(ctx, orderID, tenantID)
	if err != nil {
		return err
	}
	if o.Status != OrderStatusPending {
		return fmt.Errorf("can only reject a pending order, current status: %s", o.Status)
	}
	return s.repo.UpdateStatus(ctx, orderID, tenantID, OrderStatusRejected, nil)
}

// CompleteOrder — Admin only. Final state after shipped.
func (s *Service) CompleteOrder(ctx context.Context, tenantID, orderID string) error {
	o, err := s.getAndValidate(ctx, orderID, tenantID)
	if err != nil {
		return err
	}
	if o.Status != OrderStatusShipped {
		return fmt.Errorf("can only complete a shipped order, current status: %s", o.Status)
	}
	return s.repo.UpdateStatus(ctx, orderID, tenantID, OrderStatusCompleted, nil)
}

// ProcessOrder — Staff only.
func (s *Service) ProcessOrder(ctx context.Context, tenantID, orderID string) error {
	o, err := s.getAndValidate(ctx, orderID, tenantID)
	if err != nil {
		return err
	}
	if o.Status != OrderStatusAccepted {
		return fmt.Errorf("can only process an accepted order, current status: %s", o.Status)
	}
	return s.repo.UpdateStatus(ctx, orderID, tenantID, OrderStatusProcessing, nil)
}

// ShipOrder — Staff only.
func (s *Service) ShipOrder(ctx context.Context, tenantID, orderID string) error {
	o, err := s.getAndValidate(ctx, orderID, tenantID)
	if err != nil {
		return err
	}
	if o.Status != OrderStatusProcessing {
		return fmt.Errorf("can only ship a processing order, current status: %s", o.Status)
	}
	return s.repo.UpdateStatus(ctx, orderID, tenantID, OrderStatusShipped, nil)
}

// CancelOrder — Buyer only, only from pending.
func (s *Service) CancelOrder(ctx context.Context, tenantID, buyerID, orderID string) error {
	o, err := s.getAndValidate(ctx, orderID, tenantID)
	if err != nil {
		return err
	}
	if o.BuyerID != buyerID {
		return fmt.Errorf("order does not belong to this user")
	}
	if o.Status != OrderStatusPending {
		return fmt.Errorf("can only cancel a pending order, current status: %s", o.Status)
	}
	return s.repo.UpdateStatus(ctx, orderID, tenantID, OrderStatusCancelled, nil)
}

// ── Notes ─────────────────────────────────────────────────────────────────────

// AddNote — Buyer or field_agent only, any status.
func (s *Service) AddNote(ctx context.Context, tenantID, authorID, authorType, orderID string, req AddNoteRequest) error {
	o, err := s.getAndValidate(ctx, orderID, tenantID)
	if err != nil {
		return err
	}
	if o.BuyerID != authorID {
		return fmt.Errorf("order does not belong to this user")
	}

	note := &Note{
		AuthorID:   authorID,
		AuthorType: authorType,
		Message:    req.Message,
		CreatedAt:  time.Now().UnixMilli(),
	}
	return s.repo.AddNote(ctx, orderID, tenantID, note)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (s *Service) GetByID(ctx context.Context, tenantID, orderID string) (*Order, error) {
	return s.getAndValidate(ctx, orderID, tenantID)
}

func (s *Service) List(ctx context.Context, f ListOrdersFilter) (*ListOrdersResponse, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}

	orders, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, err
	}

	return &ListOrdersResponse{
		Orders:   orders,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

// ── Internal ──────────────────────────────────────────────────────────────────

func (s *Service) getAndValidate(ctx context.Context, orderID, tenantID string) (*Order, error) {
	o, err := s.repo.GetByID(ctx, orderID, tenantID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, fmt.Errorf("order not found")
	}
	return o, nil
}
