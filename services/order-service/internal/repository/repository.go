package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/order-service/internal/model"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error
}
