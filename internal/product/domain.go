package product

import "backend/internal/domain"

// ── Core Model ────────────────────────────────────────────────────────────────

type Product struct {
	ID                 string                     `json:"id"`
	WholesalerID       string                     `json:"wholesaler_id"` // scope: products belong to a wholesaler
	Name               string                     `json:"name"`
	Quantity           int64                      `json:"quantity"`
	Category           string                     `json:"category"` // free text, wholesaler-defined
	Unit               *domain.Unit               `json:"unit"`
	Price              int64                      `json:"price"` // in smallest currency unit (paise)
	Description        string                     `json:"description"`
	Images             []*domain.Media            `json:"images"`
	PhysicalAttributes *domain.PhysicalAttributes `json:"physical_attributes"`
	CreatedAt          int64                      `json:"created_at"`
	UpdatedAt          int64                      `json:"updated_at"`
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
