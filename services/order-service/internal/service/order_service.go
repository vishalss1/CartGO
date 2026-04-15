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
	ErrPaymentFailed         = errors.New("payment processing failed")
	ErrStockFailed           = errors.New("stock reservation failed")
	ErrPaymentNotSuccessful  = errors.New("payment not successful")
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, req *model.CreateOrderRequest) (*model.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*model.Order, error)
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Order, error)
	ListAllOrders(ctx context.Context, status string, userID string, page, limit int) ([]*model.Order, int, error)
	ConfirmPayment(ctx context.Context, id uuid.UUID) (*model.Order, error)
}

type orderService struct {
	repo            repository.OrderRepository
	inventoryClient client.InventoryClient
	paymentClient   client.PaymentClient
	productClient   client.ProductClient
	deliveryClient  client.DeliveryClient
}

func NewOrderService(
	repo repository.OrderRepository,
	invClient client.InventoryClient,
	payClient client.PaymentClient,
	prodClient client.ProductClient,
	delClient client.DeliveryClient,
) OrderService {
	return &orderService{
		repo:            repo,
		inventoryClient: invClient,
		paymentClient:   payClient,
		productClient:   prodClient,
		deliveryClient:  delClient,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, userID uuid.UUID, req *model.CreateOrderRequest) (*model.Order, error) {
	// 1. Initial Order Setup (Fetching real prices)
	order := &model.Order{
		UserID:          userID,
		Status:          model.OrderStatusPending,
		DeliveryAddress: req.DeliveryAddress,
	}

	for _, item := range req.Items {
		prod, err := s.productClient.GetProduct(ctx, item.ProductID)
		if err != nil {
			log.Printf("[OrderService] failed to fetch product %s: %v", item.ProductID, err)
			return nil, err // Propagate specific client error (e.g. "item no longer exists")
		}

		order.Items = append(order.Items, model.OrderItem{
			ProductID:    item.ProductID,
			ProductName:  prod.Name,
			Category:     prod.Category,
			Quantity:     item.Quantity,
			PricePerUnit: prod.Price,
		})
		order.TotalAmount += float64(item.Quantity) * prod.Price
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

	// 5. Commit Stock
	if err := s.inventoryClient.Commit(ctx, order.ID); err != nil {
		log.Printf("[OrderService] failed to commit stock for order %s: %v", order.ID, err)
		// This is a critical state; in a real system, this would be retried via a background worker
	}

	// 6. Create Delivery Record
	if err := s.deliveryClient.CreateDelivery(ctx, order.ID, order.DeliveryAddress); err != nil {
		log.Printf("[OrderService] failed to trigger delivery for order %s: %v", order.ID, err)
		// We proceed as order is paid and committed, but logging the delivery failure
	}

	// 7. Confirm Order
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

func (s *orderService) ListAllOrders(ctx context.Context, status string, userID string, page, limit int) ([]*model.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	return s.repo.ListAll(ctx, status, userID, limit, offset)
}

func (s *orderService) ConfirmPayment(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	// 1. Fetch Order
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrNotFound
	}

	// If already confirmed, just return
	if order.Status == model.OrderStatusConfirmed {
		return order, nil
	}

	// 2. Cross-verify Payment Status from Payment Service
	status, err := s.paymentClient.GetPaymentStatus(ctx, id)
	if err != nil {
		log.Printf("[OrderService] failed to verify payment status for order %s: %v", id, err)
		return nil, err
	}

	log.Printf("[OrderService] order %s payment status verification: received='%s'", id, status)

	if status != "SUCCESS" {
		log.Printf("[OrderService] cannot confirm order %s: payment status is '%s'", id, status)
		return nil, ErrPaymentNotSuccessful
	}

	// 3. Re-reserve Stock (compensate for release on original failure)
	for _, item := range order.Items {
		err := s.inventoryClient.Reserve(ctx, item.ProductID, order.ID, item.Quantity)
		if err != nil {
			log.Printf("[OrderService] failed to re-reserve stock for order %s, product %s: %v", order.ID, item.ProductID, err)
			return nil, fmt.Errorf("%w: %v", ErrStockFailed, err)
		}
	}

	// 4. Commit Stock
	if err := s.inventoryClient.Commit(ctx, order.ID); err != nil {
		log.Printf("[OrderService] failed to commit stock for order %s: %v", order.ID, err)
	}

	// 5. Create Delivery Record
	if err := s.deliveryClient.CreateDelivery(ctx, order.ID, order.DeliveryAddress); err != nil {
		log.Printf("[OrderService] failed to trigger delivery for order %s: %v", order.ID, err)
	}

	// 6. Update Order Status
	if err := s.repo.UpdateStatus(ctx, order.ID, model.OrderStatusConfirmed); err != nil {
		return nil, err
	}

	order.Status = model.OrderStatusConfirmed
	return order, nil
}
