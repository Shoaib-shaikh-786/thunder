package product

import (
    "backend/internal/domain"
    "time"
)

// ── Core Model ────────────────────────────────────────────────────────────────

type Product struct {
	ID               string          `json:"id" gorm:"primaryKey;type:varchar(50)"`
	TenantID         string          `json:"tenant_id" gorm:"type:varchar(50);not null;index;index:idx_tenant_sku,unique"`
	WholesalerID     string          `json:"wholesaler_id" gorm:"type:varchar(50);not null;index"`
	Name             string          `json:"name" gorm:"type:varchar(150);not null"`
	SKU              string          `json:"sku" gorm:"type:varchar(50);not null;index:idx_tenant_sku,unique"`
	Price            float64         `json:"price" gorm:"type:numeric(10,2);not null"`
	Quantity         int             `json:"quantity" gorm:"not null;default:0"`
	Category         string          `json:"category" gorm:"type:varchar(100);index"` // Indexed for fast catalog filtering
	Description      string          `json:"description" gorm:"type:text"`

	// Foreign Key for Unit Relation
	UnitID           uint            `json:"unit_id"`
	Unit             *domain.Unit    `json:"unit" gorm:"foreignKey:UnitID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
    PhysicalAttributes *domain.PhysicalAttributes `json:"physical_attributes" gorm:"type:jsonb"`
	Images           []*domain.Media `json:"images" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}




// ── Request DTOs ──────────────────────────────────────────────────────────────

type CreateProductRequest struct {
	Name               string                     `json:"name"        binding:"required"`
	Quantity           int64                      `json:"quantity"    binding:"required,min=0"`
	Category           string                     `json:"category"    binding:"required"`
	Unit               *domain.Unit               `json:"unit"`
	Price              int64                      `json:"price"       binding:"required,min=0"`
	Description        string                     `json:"description"`
	Images             []*domain.Media            `json:"images"`
	PhysicalAttributes *domain.PhysicalAttributes `json:"physical_attributes"`
}

type UpdateProductRequest struct {
	Name               *string                    `json:"name"` // all fields optional — patch semantics
	Quantity           *int64                     `json:"quantity"`
	Category           *string                    `json:"category"`
	Unit               *domain.Unit               `json:"unit"`
	Price              *int64                     `json:"price"`
	Description        *string                    `json:"description"`
	Images             []*domain.Media            `json:"images"`
	PhysicalAttributes *domain.PhysicalAttributes `json:"physical_attributes"`
}

// ── Query / Filter ────────────────────────────────────────────────────────────

type ListProductsFilter struct {
	WholesalerID string // always injected from token claims, never from request body
	Category     string // optional ?category=beverages
	Search       string // optional ?search=sugar (matches name)
	Page         int    // default 1
	PageSize     int    // default 20, max 100
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

type ListProductsResponse struct {
	Products []*Product `json:"products"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}
