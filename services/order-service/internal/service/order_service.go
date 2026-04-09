package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/order-service/internal/client"
	"github.com/vishalss1/CartGO/services/order-service/internal/model"
	"github.com/vishalss1/CartGO/services/order-service/internal/repository"
)

var (
	ErrNotFound      = errors.New("order not found")
	ErrInvalidOrder  = errors.New("invalid order request")
	ErrPaymentFailed = errors.New("payment processing failed")
	ErrStockFailed   = errors.New("stock reservation failed")
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, req *model.CreateOrderRequest) (*model.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*model.Order, error)
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Order, error)
}

type orderService struct {
	repo            repository.OrderRepository
	inventoryClient client.InventoryClient
	paymentClient   client.PaymentClient
}

func NewOrderService(
	repo repository.OrderRepository,
	invClient client.InventoryClient,
	payClient client.PaymentClient,
) OrderService {
	return &orderService{
		repo:            repo,
		inventoryClient: invClient,
		paymentClient:   payClient,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, userID uuid.UUID, req *model.CreateOrderRequest) (*model.Order, error) {
	// 1. Initial Order Setup (PENDING)
	order := &model.Order{
		UserID: userID,
		Status: model.OrderStatusPending,
	}

	for _, item := range req.Items {
		order.Items = append(order.Items, model.OrderItem{
			ProductID:    item.ProductID,
			Quantity:     item.Quantity,
			PricePerUnit: 100.0, // Mock price for now; should ideally come from Product Service
		})
		order.TotalAmount += float64(item.Quantity) * 100.0
	}

	// 2. Save Order in DB
	if err := s.repo.Create(ctx, order); err != nil {
		log.Printf("[OrderService] failed to create order in db: %v", err)
		return nil, err
	}

	// 3. Reserve Stock
	var reservedItems []model.OrderItem
	for _, item := range order.Items {
		err := s.inventoryClient.Reserve(ctx, item.ProductID, order.ID, item.Quantity)
		if err != nil {
			log.Printf("[OrderService] failed to reserve stock for order %s, product %s: %v", order.ID, item.ProductID, err)
			s.handleFailure(ctx, order, reservedItems, model.OrderStatusFailed)
			return nil, fmt.Errorf("%w: %v", ErrStockFailed, err)
		}
		reservedItems = append(reservedItems, item)
	}

	// 4. Process Payment
	if err := s.paymentClient.ProcessPayment(ctx, order.ID, order.TotalAmount); err != nil {
		log.Printf("[OrderService] payment processing failed for order %s: %v", order.ID, err)
		s.handleFailure(ctx, order, reservedItems, model.OrderStatusFailed)
		return nil, ErrPaymentFailed
	}

	// 5. Commit Stock & Confirm Order
	if err := s.inventoryClient.Commit(ctx, order.ID); err != nil {
		log.Printf("[OrderService] failed to commit stock for order %s: %v", order.ID, err)
		// This is a critical state; in a real system, this would be retried via a background worker
	}

	if err := s.repo.UpdateStatus(ctx, order.ID, model.OrderStatusConfirmed); err != nil {
		log.Printf("[OrderService] failed to update order status to confirmed for order %s: %v", order.ID, err)
		return nil, err
	}

	order.Status = model.OrderStatusConfirmed
	return order, nil
}

func (s *orderService) handleFailure(ctx context.Context, order *model.Order, reservedItems []model.OrderItem, finalStatus model.OrderStatus) {
	// Release any reserved stock
	if len(reservedItems) > 0 {
		if err := s.inventoryClient.Release(ctx, order.ID); err != nil {
			log.Printf("[OrderService] critical: failed to release stock after failure for order %s: %v", order.ID, err)
		}
	}

	// Update DB status
	if err := s.repo.UpdateStatus(ctx, order.ID, finalStatus); err != nil {
		log.Printf("[OrderService] failed to update order status to FAILED for order %s: %v", order.ID, err)
	}
}

func (s *orderService) GetOrder(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrNotFound
	}
	return order, nil
}

func (s *orderService) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Order, error) {
	return s.repo.GetByUserID(ctx, userID)
}
