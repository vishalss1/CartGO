package model

import "time"

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=3,max=255"`
	Description string  `json:"description" validate:"required"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Category    string  `json:"category" validate:"required"`
}

type UpdateProductRequest struct {
	Name        string   `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	Description string   `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	Category    string   `json:"category,omitempty"`
}

type ProductFilter struct {
	Category string   `json:"category,omitempty"`
	MinPrice *float64 `json:"min_price,omitempty"`
	MaxPrice *float64 `json:"max_price,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	Offset   int      `json:"offset,omitempty"`
}
