package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/product-service/internal/middleware"
	"github.com/vishalss1/CartGO/services/product-service/internal/model"
	"github.com/vishalss1/CartGO/services/product-service/internal/repository"
)

var (
	ErrNotFound     = errors.New("product not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal     = errors.New("internal error")
)

type ProductService interface {
	CreateProduct(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*model.Product, error)
	ListProducts(ctx context.Context, filter *model.ProductFilter) ([]*model.Product, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, req *model.UpdateProductRequest) (*model.Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error) {
	if req.Name == "" || req.Price <= 0 {
		return nil, ErrInvalidInput
	}

	p := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
	}

	res, err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	// ── Audit Logging ──
	log.Printf("[AUDIT] user=%s role=%s action=CREATE_PRODUCT entityId=%s timestamp=%s", 
		middleware.GetUserID(ctx), middleware.GetRole(ctx), res.ID, time.Now().Format(time.RFC3339))

	return res, nil
}

func (s *productService) GetProduct(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}
	return p, nil
}

func (s *productService) ListProducts(ctx context.Context, filter *model.ProductFilter) ([]*model.Product, error) {
	// Defaults handled in handler
	products, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	// ── Response Sanitization (Backend Guard) ──
	validProducts := make([]*model.Product, 0, len(products))
	for _, p := range products {
		if p.Name == "" || p.Price <= 0 {
			log.Printf("[ProductService] WARN: malformed product record found in DB (ID: %s, Name: %q, Price: %.2f). Filtering from response.", p.ID, p.Name, p.Price)
			continue
		}
		validProducts = append(validProducts, p)
	}

	return validProducts, nil
}

func (s *productService) UpdateProduct(ctx context.Context, id uuid.UUID, req *model.UpdateProductRequest) (*model.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	// Apply partial updates
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.Category != "" {
		p.Category = req.Category
	}

	res, err := s.repo.Update(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	// ── Audit Logging ──
	log.Printf("[AUDIT] user=%s role=%s action=UPDATE_PRODUCT entityId=%s timestamp=%s", 
		middleware.GetUserID(ctx), middleware.GetRole(ctx), res.ID, time.Now().Format(time.RFC3339))

	return res, nil
}

func (s *productService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	// ── Audit Logging ──
	log.Printf("[AUDIT] user=%s role=%s action=DELETE_PRODUCT entityId=%s timestamp=%s", 
		middleware.GetUserID(ctx), middleware.GetRole(ctx), id, time.Now().Format(time.RFC3339))

	return nil
}
