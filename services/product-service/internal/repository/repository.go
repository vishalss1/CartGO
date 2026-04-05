package repository

import (
	"context"

	"github.com/vishalss1/CartGO/services/product-service/internal/model"
)

type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) (*model.Product, error)
	GetByID(ctx context.Context, id string) (*model.Product, error)
	List(ctx context.Context, filter *model.ProductFilter) ([]*model.Product, error)
	Update(ctx context.Context, product *model.Product) (*model.Product, error)
	Delete(ctx context.Context, id string) error
}
