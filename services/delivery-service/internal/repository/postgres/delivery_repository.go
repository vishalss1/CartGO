package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/vishalss1/CartGO/pkg/util"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/model"
)

type PostgresDeliveryRepository struct {
	db *sql.DB
}

func NewPostgresDeliveryRepository(db *sql.DB) *PostgresDeliveryRepository {
	return &PostgresDeliveryRepository{db: db}
}

func (r *PostgresDeliveryRepository) Create(ctx context.Context, d *model.Delivery) error {
	if d.ID == uuid.Nil {
		d.ID = util.GenerateUUID()
	}

	query := `
		INSERT INTO deliveries (id, order_id, status, delivery_address)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, d.ID, d.OrderID, d.Status, d.DeliveryAddress).
		Scan(&d.CreatedAt, &d.UpdatedAt)

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return fmt.Errorf("delivery already exists for order %s", d.OrderID)
		}
		return fmt.Errorf("failed to create delivery: %v", err)
	}

	return nil
}

func (r *PostgresDeliveryRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.DeliveryStatus, partnerID *uuid.UUID) error {
	query := `
		UPDATE deliveries
		SET status = $1, delivery_person_id = COALESCE($2, delivery_person_id), updated_at = CURRENT_TIMESTAMP
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, status, partnerID, id)
	if err != nil {
		return fmt.Errorf("failed to update delivery status: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("delivery not found")
	}

	return nil
}

func (r *PostgresDeliveryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Delivery, error) {
	query := `
		SELECT id, order_id, status, delivery_address, delivery_person_id, created_at, updated_at
		FROM deliveries
		WHERE id = $1`

	d := &model.Delivery{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.OrderID, &d.Status, &d.DeliveryAddress, &d.DeliveryPersonID, &d.CreatedAt, &d.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("delivery not found")
		}
		return nil, fmt.Errorf("failed to get delivery: %v", err)
	}

	return d, nil
}

func (r *PostgresDeliveryRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Delivery, error) {
	query := `
		SELECT id, order_id, status, delivery_address, delivery_person_id, created_at, updated_at
		FROM deliveries
		WHERE order_id = $1`

	d := &model.Delivery{}
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&d.ID, &d.OrderID, &d.Status, &d.DeliveryAddress, &d.DeliveryPersonID, &d.CreatedAt, &d.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("delivery not found")
		}
		return nil, fmt.Errorf("failed to get delivery by order_id: %v", err)
	}

	return d, nil
}

func (r *PostgresDeliveryRepository) ListByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*model.Delivery, error) {
	query := `
		SELECT id, order_id, status, delivery_address, delivery_person_id, created_at, updated_at
		FROM deliveries
		WHERE delivery_person_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list deliveries: %v", err)
	}
	defer rows.Close()

	var deliveries []*model.Delivery
	for rows.Next() {
		d := &model.Delivery{}
		err := rows.Scan(&d.ID, &d.OrderID, &d.Status, &d.DeliveryAddress, &d.DeliveryPersonID, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}

	return deliveries, nil
}
