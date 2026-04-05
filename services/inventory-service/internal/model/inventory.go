package model

import (
	"time"

	"github.com/google/uuid"
)

type Inventory struct {
	ProductID string    `json:"product_id"`
	Stock     int       `json:"stock"`
	Reserved  int       `json:"reserved"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Reservation struct {
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Status    string    `json:"status"` // PENDING, COMMITTED, RELEASED
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateStockRequest struct {
	TotalStock int `json:"total_stock" validate:"required,min=0"`
}

type ReserveRequest struct {
	OrderID  uuid.UUID `json:"order_id" validate:"required"`
	Quantity int       `json:"quantity" validate:"required,gt=0"`
}

type IdempotentRequest struct {
	OrderID uuid.UUID `json:"order_id" validate:"required"`
}

type InventoryResponse struct {
	ProductID      string `json:"product_id"`
	AvailableStock int    `json:"available_stock"`
	TotalStock     int    `json:"total_stock"`
	ReservedStock  int    `json:"reserved_stock"`
}

func (i *Inventory) ToResponse() *InventoryResponse {
	return &InventoryResponse{
		ProductID:      i.ProductID,
		AvailableStock: i.Stock - i.Reserved,
		TotalStock:     i.Stock,
		ReservedStock:  i.Reserved,
	}
}
