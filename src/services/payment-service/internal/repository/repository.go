package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/payment-service/internal/model"
)

var (
	ErrNotFound = errors.New("payment not found")
	ErrConflict = errors.New("payment for this order already exists")
)

type PaymentRepository interface {
	Save(ctx context.Context, payment *model.Payment) (*model.Payment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Payment, error)
}
