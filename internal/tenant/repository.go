package tenant

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Tenant, error) {
	t := &Tenant{}
	err := r.db.QueryRow(ctx, `
        SELECT id, company_name, business_type, branding_config, is_active, created_at, updated_at
        FROM tenants
        WHERE id = $1
          AND is_active = true
    `, id).Scan(&t.ID, &t.CompanyName, &t.BusinessType, &t.BrandingConfig, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get tenant: %w", err)
	}
	return t, nil
}
