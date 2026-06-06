package product

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ── Write ─────────────────────────────────────────────────────────────────────

func (r *Repository) Create(ctx context.Context, p *Product) error {
	imagesJSON, err := json.Marshal(p.Images)
	if err != nil {
		return fmt.Errorf("marshal images: %w", err)
	}
	physJSON, err := json.Marshal(p.PhysicalAttributes)
	if err != nil {
		return fmt.Errorf("marshal physical_attributes: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO products (
			id, wholesaler_id, name, quantity, category,
			unit, price, description, images, physical_attributes,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12
		)
	`,
		p.ID, p.WholesalerID, p.Name, p.Quantity, p.Category,
		p.Unit, p.Price, p.Description, imagesJSON, physJSON,
		p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *Repository) Update(ctx context.Context, id, wholesalerID string, req UpdateProductRequest) error {
	// Build query dynamically — only update provided fields
	setClauses := []string{}
	args := []any{}
	argIdx := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Quantity != nil {
		setClauses = append(setClauses, fmt.Sprintf("quantity = $%d", argIdx))
		args = append(args, *req.Quantity)
		argIdx++
	}
	if req.Category != nil {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *req.Category)
		argIdx++
	}
	if req.Unit != nil {
		setClauses = append(setClauses, fmt.Sprintf("unit = $%d", argIdx))
		args = append(args, *req.Unit)
		argIdx++
	}
	if req.Price != nil {
		setClauses = append(setClauses, fmt.Sprintf("price = $%d", argIdx))
		args = append(args, *req.Price)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Images != nil {
		imagesJSON, err := json.Marshal(req.Images)
		if err != nil {
			return fmt.Errorf("marshal images: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("images = $%d", argIdx))
		args = append(args, imagesJSON)
		argIdx++
	}
	if req.PhysicalAttributes != nil {
		physJSON, err := json.Marshal(req.PhysicalAttributes)
		if err != nil {
			return fmt.Errorf("marshal physical_attributes: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("physical_attributes = $%d", argIdx))
		args = append(args, physJSON)
		argIdx++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Always update updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now().UnixMilli())
	argIdx++

	// WHERE id = $N AND wholesaler_id = $N+1 (prevents cross-wholesaler edits)
	args = append(args, id, wholesalerID)
	query := fmt.Sprintf(
		"UPDATE products SET %s WHERE id = $%d AND wholesaler_id = $%d",
		strings.Join(setClauses, ", "), argIdx, argIdx+1,
	)

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found or not owned by wholesaler")
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id, wholesalerID string) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM products WHERE id = $1 AND wholesaler_id = $2`,
		id, wholesalerID,
	)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found or not owned by wholesaler")
	}
	return nil
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *Repository) GetByID(ctx context.Context, id, wholesalerID string) (*Product, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, wholesaler_id, name, quantity, category,
		       unit, price, description, images, physical_attributes,
		       created_at, updated_at
		FROM products
		WHERE id = $1 AND wholesaler_id = $2
	`, id, wholesalerID)

	return scanProduct(row)
}

func (r *Repository) List(ctx context.Context, f ListProductsFilter) ([]*Product, int64, error) {
	conditions := []string{"wholesaler_id = $1"}
	args := []any{f.WholesalerID}
	argIdx := 2

	if f.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, f.Category)
		argIdx++
	}
	if f.Search != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+f.Search+"%")
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Total count
	var total int64
	err := r.db.QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM products %s", where),
		args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	// Paginated results
	offset := (f.Page - 1) * f.PageSize
	args = append(args, f.PageSize, offset)
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, wholesaler_id, name, quantity, category,
		       unit, price, description, images, physical_attributes,
		       created_at, updated_at
		FROM products %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}

	return products, total, nil
}

// ── Scanner ───────────────────────────────────────────────────────────────────

type scannable interface {
	Scan(dest ...any) error
}

func scanProduct(row scannable) (*Product, error) {
	p := &Product{}
	var unitStr *string
	var imagesJSON []byte
	var physJSON []byte

	err := row.Scan(
		&p.ID, &p.WholesalerID, &p.Name, &p.Quantity, &p.Category,
		&unitStr, &p.Price, &p.Description, &imagesJSON, &physJSON,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan product: %w", err)
	}

	if unitStr != nil {
		u := domain.Unit(*unitStr)
		p.Unit = &u
	}
	if len(imagesJSON) > 0 {
		if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
			return nil, fmt.Errorf("unmarshal images: %w", err)
		}
	}
	if len(physJSON) > 0 {
		if err := json.Unmarshal(physJSON, &p.PhysicalAttributes); err != nil {
			return nil, fmt.Errorf("unmarshal physical_attributes: %w", err)
		}
	}

	return p, nil
}
