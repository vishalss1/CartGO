package service

import (
	"context"

	"github.com/vishalss1/CartGO/services/product-service/internal/model"
	"github.com/vishalss1/CartGO/services/product-service/internal/repository"
)

type ProductService interface {
	CreateProduct(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error)
	GetProduct(ctx context.Context, id string) (*model.Product, error)
	ListProducts(ctx context.Context, filter *model.ProductFilter) ([]*model.Product, error)
	UpdateProduct(ctx context.Context, id string, req *model.UpdateProductRequest) (*model.Product, error)
	DeleteProduct(ctx context.Context, id string) error
	AdjustStock(ctx context.Context, id string, delta int) error
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error) {
	p := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		Stock:       req.Stock,
	}

	return s.repo.Create(ctx, p)
}

func (s *productService) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *productService) ListProducts(ctx context.Context, filter *model.ProductFilter) ([]*model.Product, error) {
	return s.repo.List(ctx, filter)
}

func (s *productService) UpdateProduct(ctx context.Context, id string, req *model.UpdateProductRequest) (*model.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

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
	if req.Stock != nil {
		p.Stock = *req.Stock
	}

	return s.repo.Update(ctx, p)
}

func (s *productService) DeleteProduct(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *productService) AdjustStock(ctx context.Context, id string, delta int) error {
	return s.repo.UpdateStock(ctx, id, delta)
}
