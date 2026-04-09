package service

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"log"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/payment-service/internal/model"
	"github.com/vishalss1/CartGO/services/payment-service/internal/repository/postgres"
)

var (
	ErrInternal        = errors.New("internal server error")
	ErrAmountMismatch  = errors.New("amount mismatch for same order_id")
)

type PaymentService interface {
	ProcessPayment(ctx context.Context, orderID uuid.UUID, amount float64) (*model.Payment, error)
}

type paymentService struct {
	repo   *postgres.PostgresPaymentRepository
}

func NewPaymentService(repo *postgres.PostgresPaymentRepository) PaymentService {
	return &paymentService{
		repo:   repo,
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

	// Idempotency: verify amount matches if it's an existing record
	if savedPayment.ID != uuid.Nil && savedPayment.Amount != amount {
		log.Printf("[PaymentService] payment amount mismatch for existing order %s: existing=%.2f, requested=%.2f", orderID, savedPayment.Amount, amount)
		return nil, ErrAmountMismatch
	}

	return savedPayment, nil
}

func (s *paymentService) isSuccessful(orderID string) bool {
	h := fnv.New32a()
	h.Write([]byte(orderID))
	return h.Sum32()%100 < 85
}
