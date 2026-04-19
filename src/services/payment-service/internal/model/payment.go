package model

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusSuccess PaymentStatus = "SUCCESS"
	PaymentStatusFailure PaymentStatus = "FAILURE"
)

type Payment struct {
	ID        uuid.UUID     `json:"id"`
	OrderID   uuid.UUID     `json:"order_id"`
	Amount    float64       `json:"amount"`
	Status    PaymentStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}

type PaymentRequest struct {
	OrderID uuid.UUID `json:"order_id" validate:"required"`
	Amount  float64   `json:"amount" validate:"required,gt=0"`
}

type PaymentResponse struct {
	PaymentID uuid.UUID `json:"payment_id"`
	OrderID   uuid.UUID `json:"order_id"`
	Status    string    `json:"status"`
	Amount    float64   `json:"amount"`
}
