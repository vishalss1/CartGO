package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/payment-service/internal/client"
	"github.com/vishalss1/CartGO/services/payment-service/internal/model"
	"github.com/vishalss1/CartGO/services/payment-service/internal/repository/postgres"
)

var (
	ErrInternal        = errors.New("internal server error")
	ErrAmountMismatch  = errors.New("amount mismatch for same order_id")
)

type PaymentService interface {
	ProcessPayment(ctx context.Context, orderID uuid.UUID, amount float64) (*model.Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Payment, error)
}

type paymentService struct {
	repo        *postgres.PostgresPaymentRepository
	orderClient client.OrderClient
}

func NewPaymentService(repo *postgres.PostgresPaymentRepository, orderClient client.OrderClient) PaymentService {
	return &paymentService{
		repo:        repo,
		orderClient: orderClient,
	}
}

func (s *paymentService) ProcessPayment(ctx context.Context, orderID uuid.UUID, amount float64) (*model.Payment, error) {
	// Deterministic decision (85% success rate)
	status := model.PaymentStatusSuccess
	if !s.isSuccessful(orderID.String()) {
		status = model.PaymentStatusFailure
	}

	payment := &model.Payment{
		OrderID: orderID,
		Amount:  amount,
		Status:  status,
	}

	log.Printf("[PaymentService] processing payment: order_id=%s, amount=%.2f, decision=%s", orderID, amount, status)

	savedPayment, err := s.repo.Save(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	if savedPayment.ID != uuid.Nil && savedPayment.Amount != amount {
		log.Printf("[PaymentService] payment amount mismatch for existing order %s: existing=%.2f, requested=%.2f", orderID, savedPayment.Amount, amount)
		return nil, ErrAmountMismatch
	}

	// Proactive Notification to Order Service for Success
	if savedPayment.Status == model.PaymentStatusSuccess {
		go func() {
			log.Printf("[PaymentService] Proactively notifying Order Service for confirmation of order %s", orderID)
			err := s.orderClient.ConfirmOrder(context.Background(), orderID)
			if err != nil {
				log.Printf("[PaymentService] Warning: Failed to notify Order Service for order %s: %v. Fallback to frontend/polling resolution.", orderID, err)
			} else {
				log.Printf("[PaymentService] Success: Order Service notified and confirmation triggered for order %s", orderID)
			}
		}()
	}

	return savedPayment, nil
}

func (s *paymentService) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Payment, error) {
	return s.repo.GetByOrderID(ctx, orderID)
}

func (s *paymentService) isSuccessful(orderID string) bool {
	// Seed the random number generator so each attempt is non-deterministic
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(100) < 85
}
