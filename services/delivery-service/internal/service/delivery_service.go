package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/model"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/repository/postgres"
)

type DeliveryService interface {
	CreateDelivery(ctx context.Context, orderID uuid.UUID, address string) (*model.Delivery, error)
	UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status model.DeliveryStatus, partnerID *uuid.UUID) error
	GetDelivery(ctx context.Context, id uuid.UUID) (*model.Delivery, error)
	GetDeliveryByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Delivery, error)
	ListDeliveriesByPartner(ctx context.Context, partnerID uuid.UUID) ([]*model.Delivery, error)
}

type deliveryService struct {
	repo   *postgres.PostgresDeliveryRepository
	logger *slog.Logger
}

func NewDeliveryService(repo *postgres.PostgresDeliveryRepository, logger *slog.Logger) DeliveryService {
	return &deliveryService{
		repo:   repo,
		logger: logger,
	}
}

func (s *deliveryService) CreateDelivery(ctx context.Context, orderID uuid.UUID, address string) (*model.Delivery, error) {
	delivery := &model.Delivery{
		OrderID:         orderID,
		Status:          model.DeliveryStatusPending,
		DeliveryAddress: address,
	}

	err := s.repo.Create(ctx, delivery)
	if err != nil {
		s.logger.Error("failed to create delivery", "order_id", orderID, "error", err)
		return nil, err
	}

	s.logger.Info("delivery created", "delivery_id", delivery.ID, "order_id", orderID)
	return delivery, nil
}

func (s *deliveryService) UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status model.DeliveryStatus, partnerID *uuid.UUID) error {
	err := s.repo.UpdateStatus(ctx, id, status, partnerID)
	if err != nil {
		s.logger.Error("failed to update delivery status", "delivery_id", id, "status", status, "error", err)
		return err
	}

	s.logger.Info("delivery status updated", "delivery_id", id, "status", status, "partner_id", partnerID)
	return nil
}

func (s *deliveryService) GetDelivery(ctx context.Context, id uuid.UUID) (*model.Delivery, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *deliveryService) GetDeliveryByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Delivery, error) {
	return s.repo.GetByOrderID(ctx, orderID)
}

func (s *deliveryService) ListDeliveriesByPartner(ctx context.Context, partnerID uuid.UUID) ([]*model.Delivery, error) {
	return s.repo.ListByPartnerID(ctx, partnerID)
}
