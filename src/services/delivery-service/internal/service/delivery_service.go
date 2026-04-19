package service

import (
	"context"
	"log"

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
	ListAvailableDeliveries(ctx context.Context) ([]*model.Delivery, error)
}

type deliveryService struct {
	repo   *postgres.PostgresDeliveryRepository
}

func NewDeliveryService(repo *postgres.PostgresDeliveryRepository) DeliveryService {
	return &deliveryService{
		repo:   repo,
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
		log.Printf("[DeliveryService] failed to create delivery for order %s: %v", orderID, err)
		return nil, err
	}

	log.Printf("[DeliveryService] delivery created: %s, order_id: %s", delivery.ID, orderID)
	return delivery, nil
}

func (s *deliveryService) UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status model.DeliveryStatus, partnerID *uuid.UUID) error {
	err := s.repo.UpdateStatus(ctx, id, status, partnerID)
	if err != nil {
		log.Printf("[DeliveryService] failed to update delivery %s status to %s: %v", id, status, err)
		return err
	}

	log.Printf("[DeliveryService] delivery %s status updated to %s (partner: %v)", id, status, partnerID)
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

func (s *deliveryService) ListAvailableDeliveries(ctx context.Context) ([]*model.Delivery, error) {
	return s.repo.ListUnassigned(ctx)
}
