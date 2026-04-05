package repository

import (
	"context"

	"github.com/vishalss1/CartGO/services/inventory-service/internal/model"
)

type InventoryRepository interface {
	GetByProductID(ctx context.Context, productID string) (*model.Inventory, error)
	Upsert(ctx context.Context, inv *model.Inventory) (*model.Inventory, error)
	
	// Reservation management
	GetReservation(ctx context.Context, orderID string) (*model.Reservation, error)
	
	// Atomic operations with idempotency support
	Reserve(ctx context.Context, productID string, orderID string, quantity int, currentVersion int) error
	Release(ctx context.Context, orderID string, inventoryVersion int) error
	Commit(ctx context.Context, orderID string, inventoryVersion int) error
	
	// Direct stock update (for warehouse staff)
	UpdateStock(ctx context.Context, productID string, totalStock int, currentVersion int) error
}
