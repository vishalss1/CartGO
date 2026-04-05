package repository

import (
	"context"

	"github.com/vishalss1/CartGO/services/inventory-service/internal/model"
)

type InventoryRepository interface {
	GetByProductID(ctx context.Context, productID string) (*model.Inventory, error)
	Upsert(ctx context.Context, inv *model.Inventory) (*model.Inventory, error)
	
	// Atomic operations with internal checks
	Reserve(ctx context.Context, productID string, quantity int, currentVersion int) error
	Release(ctx context.Context, productID string, quantity int, currentVersion int) error
	Commit(ctx context.Context, productID string, quantity int, currentVersion int) error
	
	// Direct stock update (for warehouse staff)
	UpdateStock(ctx context.Context, productID string, totalStock int, currentVersion int) error
}
