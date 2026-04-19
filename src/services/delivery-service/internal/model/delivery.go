package model

import (
	"time"

	"github.com/google/uuid"
)

type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "PENDING"
	DeliveryStatusPickedUp  DeliveryStatus = "PICKED_UP"
	DeliveryStatusDelivered DeliveryStatus = "DELIVERED"
	DeliveryStatusCancelled DeliveryStatus = "CANCELLED"
)

type Delivery struct {
	ID               uuid.UUID      `json:"id"`
	OrderID          uuid.UUID      `json:"order_id"`
	Status           DeliveryStatus `json:"status"`
	DeliveryAddress  string         `json:"delivery_address"`
	DeliveryPersonID *uuid.UUID     `json:"delivery_person_id,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type CreateDeliveryRequest struct {
	OrderID         uuid.UUID `json:"order_id" validate:"required"`
	DeliveryAddress string    `json:"delivery_address" validate:"required"`
}

type UpdateDeliveryStatusRequest struct {
	Status           DeliveryStatus `json:"status" validate:"required"`
	DeliveryPersonID *uuid.UUID     `json:"delivery_person_id,omitempty"`
}
