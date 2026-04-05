package model

import "time"

type Inventory struct {
	ProductID string    `json:"product_id"`
	Stock     int       `json:"stock"`
	Reserved  int       `json:"reserved"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateStockRequest struct {
	TotalStock int `json:"total_stock" validate:"required,min=0"`
}

type StockOperationRequest struct {
	Quantity int `json:"quantity" validate:"required,gt=0"`
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
