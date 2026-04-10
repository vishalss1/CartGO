package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/inventory-service/internal/model"
)

type InventoryRepository interface {
	GetByProductID(ctx context.Context, productID uuid.UUID) (*model.Inventory, error)
	Upsert(ctx context.Context, inv *model.Inventory) (*model.Inventory, error)
	
	// Reservation management
	GetReservations(ctx context.Context, orderID uuid.UUID) ([]*model.Reservation, error)
	
	// Atomic operations with idempotency support
	Reserve(ctx context.Context, productID uuid.UUID, orderID uuid.UUID, quantity int, currentVersion int) error
	Release(ctx context.Context, orderID uuid.UUID) error
	Commit(ctx context.Context, orderID uuid.UUID) error
	
	// Direct stock update (for warehouse staff)
	UpdateStock(ctx context.Context, productID uuid.UUID, totalStock int, currentVersion int) error
}
