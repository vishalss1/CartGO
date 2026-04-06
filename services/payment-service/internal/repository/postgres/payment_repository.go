package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/vishalss1/CartGO/services/payment-service/internal/model"
)

type PostgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) Save(ctx context.Context, payment *model.Payment) (*model.Payment, error) {
	query := `
		INSERT INTO payments (order_id, amount, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		payment.OrderID, payment.Amount, payment.Status).
		Scan(&payment.ID, &payment.CreatedAt)

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
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
