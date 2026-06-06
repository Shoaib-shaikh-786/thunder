package product

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

func (s *Service) Create(ctx context.Context, wholesalerID string, req CreateProductRequest) (*Product, error) {
	p := &Product{
		ID:                 uuid.New().String(),
		WholesalerID:       wholesalerID, // always from token, never from request
		Name:               req.Name,
		Quantity:           req.Quantity,
		Category:           req.Category,
		Unit:               req.Unit,
		Price:              req.Price,
		Description:        req.Description,
		Images:             req.Images,
		PhysicalAttributes: req.PhysicalAttributes,
		CreatedAt:          time.Now().UnixMilli(),
		UpdatedAt:          time.Now().UnixMilli(),
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	return p, nil
}

func (s *Service) Update(ctx context.Context, wholesalerID, productID string, req UpdateProductRequest) error {
	// wholesaler_id check is enforced in the DB query (WHERE wholesaler_id = ?)
	return s.repo.Update(ctx, productID, wholesalerID, req)
}

func (s *Service) Delete(ctx context.Context, wholesalerID, productID string) error {
	return s.repo.Delete(ctx, productID, wholesalerID)
}

func (s *Service) GetByID(ctx context.Context, wholesalerID, productID string) (*Product, error) {
	p, err := s.repo.GetByID(ctx, productID, wholesalerID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("product not found")
	}
	return p, nil
}

func (s *Service) List(ctx context.Context, f ListProductsFilter) (*ListProductsResponse, error) {
	// Sanitize pagination
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}

	products, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, err
	}

	return &ListProductsResponse{
		Products: products,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}
