package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/pkg/util"
	"github.com/vishalss1/CartGO/services/payment-service/internal/model"
)

type PostgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) Save(ctx context.Context, payment *model.Payment) (*model.Payment, error) {
	if payment.ID == uuid.Nil {
		payment.ID = util.GenerateUUID()
	}

	query := `
		INSERT INTO payments (id, order_id, amount, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (order_id) DO UPDATE 
		SET status = EXCLUDED.status
		WHERE payments.amount = EXCLUDED.amount
		RETURNING id, amount, created_at`

	err := r.db.QueryRowContext(ctx, query,
		payment.ID, payment.OrderID, payment.Amount, payment.Status).
		Scan(&payment.ID, &payment.Amount, &payment.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// This means there was a conflict AND the amount mismatched
			// Return the existing record so the service throws ErrAmountMismatch
			return r.GetByOrderID(ctx, payment.OrderID)
		}
		return nil, fmt.Errorf("failed to save payment: %v", err)
	}

	return payment, nil
}

func (r *PostgresPaymentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Payment, error) {
	query := `
		SELECT id, order_id, amount, status, created_at
		FROM payments
		WHERE order_id = $1`

	payment := &model.Payment{}
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID, &payment.OrderID, &payment.Amount, &payment.Status, &payment.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment by order_id: %v", err)
	}

	return payment, nil
}
