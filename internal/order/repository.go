package order

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ── Write ─────────────────────────────────────────────────────────────────────

func (r *Repository) Create(ctx context.Context, o *Order) error {
	itemsJSON, err := json.Marshal(o.Items)
	if err != nil {
		return fmt.Errorf("marshal items: %w", err)
	}
	addrJSON, err := json.Marshal(o.ShippingAddress)
	if err != nil {
		return fmt.Errorf("marshal address: %w", err)
	}
	notesJSON, err := json.Marshal(o.Notes)
	if err != nil {
		return fmt.Errorf("marshal notes: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO orders (
			id, tenant_id, buyer_id, placed_by_id, placed_by_type,
			status, items, order_value, shipping_address, etd, notes,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11,
			$12, $13
		)
	`,
		o.ID, o.TenantID, o.BuyerID, o.PlacedByID, o.PlacedByType,
		o.Status, itemsJSON, o.OrderValue, addrJSON, o.ETD, notesJSON,
		o.CreatedAt, o.UpdatedAt,
	)
	return err
}

func (r *Repository) UpdateStatus(ctx context.Context, id, tenantID string, status OrderStatus, etd *time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE orders
		SET status = $1, etd = $2, updated_at = $3
		WHERE id = $4 AND tenant_id = $5
	`, status, etd, time.Now().UnixMilli(), id, tenantID)
	return err
}

func (r *Repository) AddNote(ctx context.Context, id, tenantID string, note *Note) error {
	noteJSON, err := json.Marshal(note)
	if err != nil {
		return fmt.Errorf("marshal note: %w", err)
	}

	// Append to existing JSONB array atomically
	_, err = r.db.Exec(ctx, `
		UPDATE orders
		SET notes = notes || $1::jsonb, updated_at = $2
		WHERE id = $3 AND tenant_id = $4
	`, fmt.Sprintf("[%s]", noteJSON), time.Now().UnixMilli(), id, tenantID)
	return err
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *Repository) GetByID(ctx context.Context, id, tenantID string) (*Order, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, buyer_id, placed_by_id, placed_by_type,
		       status, items, order_value, shipping_address, etd, notes,
		       created_at, updated_at
		FROM orders
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)

	return scanOrder(row)
}

func (r *Repository) List(ctx context.Context, f ListOrdersFilter) ([]*Order, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []any{f.TenantID}
	argIdx := 2

	if f.BuyerID != "" {
		conditions = append(conditions, fmt.Sprintf("buyer_id = $%d", argIdx))
		args = append(args, f.BuyerID)
		argIdx++
	}
	if f.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, f.Status)
		argIdx++
	}
	if f.PlacedByID != "" {
		conditions = append(conditions, fmt.Sprintf("placed_by_id = $%d", argIdx))
		args = append(args, f.PlacedByID)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	var total int64
	err := r.db.QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM orders %s", where),
		args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	offset := (f.Page - 1) * f.PageSize
	args = append(args, f.PageSize, offset)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, buyer_id, placed_by_id, placed_by_type,
		       status, items, order_value, shipping_address, etd, notes,
		       created_at, updated_at
		FROM orders %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

// GetUserAddress fetches the saved address from the users table.
func (r *Repository) GetUserAddress(ctx context.Context, userID string) (*domain.Address, error) {
	var addrJSON []byte
	err := r.db.QueryRow(ctx,
		`SELECT address FROM users WHERE id = $1`, userID,
	).Scan(&addrJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user address: %w", err)
	}
	if len(addrJSON) == 0 {
		return nil, nil
	}
	var addr domain.Address
	if err := json.Unmarshal(addrJSON, &addr); err != nil {
		return nil, fmt.Errorf("unmarshal address: %w", err)
	}
	return &addr, nil
}

// GetProductSnapshot fetches name, price, unit from products for snapshotting.
func (r *Repository) GetProductSnapshot(ctx context.Context, productID, tenantID string) (*OrderItem, error) {
	item := &OrderItem{}
	var unitStr *string
	err := r.db.QueryRow(ctx, `
		SELECT id, name, price, unit
		FROM products
		WHERE id = $1 AND tenant_id = $2
	`, productID, tenantID).Scan(&item.ProductID, &item.Name, &item.Price, &unitStr)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product %s not found", productID)
		}
		return nil, fmt.Errorf("get product snapshot: %w", err)
	}
	if unitStr != nil {
		item.Unit = *unitStr
	}
	return item, nil
}

// ── Scanner ───────────────────────────────────────────────────────────────────

type scannable interface {
	Scan(dest ...any) error
}

func scanOrder(row scannable) (*Order, error) {
	o := &Order{}
	var itemsJSON, addrJSON, notesJSON []byte

	err := row.Scan(
		&o.ID, &o.TenantID, &o.BuyerID, &o.PlacedByID, &o.PlacedByType,
		&o.Status, &itemsJSON, &o.OrderValue, &addrJSON, &o.ETD, &notesJSON,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan order: %w", err)
	}

	if len(itemsJSON) > 0 {
		if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
			return nil, fmt.Errorf("unmarshal items: %w", err)
		}
	}
	if len(addrJSON) > 0 {
		o.ShippingAddress = &domain.Address{}
		if err := json.Unmarshal(addrJSON, o.ShippingAddress); err != nil {
			return nil, fmt.Errorf("unmarshal address: %w", err)
		}
	}
	if len(notesJSON) > 0 {
		if err := json.Unmarshal(notesJSON, &o.Notes); err != nil {
			return nil, fmt.Errorf("unmarshal notes: %w", err)
		}
	}

	return o, nil
}
