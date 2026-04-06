package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/model"
)

var (
	ErrNotFound = errors.New("delivery not found")
	ErrConflict = errors.New("delivery already exists for this order")
)

type DeliveryRepository interface {
	Create(ctx context.Context, delivery *model.Delivery) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.DeliveryStatus, partnerID *uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Delivery, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Delivery, error)
	ListByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*model.Delivery, error)
}
