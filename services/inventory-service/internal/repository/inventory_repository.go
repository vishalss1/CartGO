package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/inventory-service/internal/model"
)

var (
	ErrConflict    = errors.New("inventory record was modified by another process")
	ErrNotFound    = errors.New("inventory record not found")
	ErrIdempotent  = errors.New("operation already performed")
	ErrInvalidState = errors.New("invalid reservation state for this operation")
)

type PostgresInventoryRepository struct {
	db *sql.DB
}

func NewPostgresInventoryRepository(db *sql.DB) *PostgresInventoryRepository {
	return &PostgresInventoryRepository{db: db}
}

func (r *PostgresInventoryRepository) GetByProductID(ctx context.Context, productID string) (*model.Inventory, error) {
	query := `
		SELECT product_id, stock, reserved, version, created_at, updated_at
		FROM inventory
		WHERE product_id = $1`

	inv := &model.Inventory{}
	err := r.db.QueryRowContext(ctx, query, productID).
		Scan(&inv.ProductID, &inv.Stock, &inv.Reserved, &inv.Version, &inv.CreatedAt, &inv.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("could not get inventory: %v", err)
	}

	return inv, nil
}

func (r *PostgresInventoryRepository) Upsert(ctx context.Context, inv *model.Inventory) (*model.Inventory, error) {
	query := `
		INSERT INTO inventory (product_id, stock, reserved, version)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (product_id) DO UPDATE SET stock = EXCLUDED.stock, updated_at = NOW()
		RETURNING version, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, inv.ProductID, inv.Stock, inv.Reserved).
		Scan(&inv.Version, &inv.CreatedAt, &inv.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("could not upsert inventory: %v", err)
	}

	return inv, nil
}

func (r *PostgresInventoryRepository) GetReservation(ctx context.Context, orderID string) (*model.Reservation, error) {
	query := `
		SELECT order_id, product_id, quantity, status, created_at, updated_at
		FROM reservations
		WHERE order_id = $1`

	res := &model.Reservation{}
	err := r.db.QueryRowContext(ctx, query, orderID).
		Scan(&res.OrderID, &res.ProductID, &res.Quantity, &res.Status, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("could not get reservation: %v", err)
	}

	return res, nil
}

func (r *PostgresInventoryRepository) Reserve(ctx context.Context, productID string, orderID string, quantity int, currentVersion int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check for existing reservation (Idempotency)
	var existingStatus string
	err = tx.QueryRowContext(ctx, "SELECT status FROM reservations WHERE order_id = $1", orderID).Scan(&existingStatus)
	if err == nil {
		return ErrIdempotent
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Update inventory with stock check and optimistic locking
	queryInv := `
		UPDATE inventory
		SET reserved = reserved + $1, version = version + 1, updated_at = NOW()
		WHERE product_id = $2 AND (stock - reserved) >= $1 AND version = $3`

	res, err := tx.ExecContext(ctx, queryInv, quantity, productID, currentVersion)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrConflict
	}

	// Create reservation record
	queryRes := `
		INSERT INTO reservations (order_id, product_id, quantity, status)
		VALUES ($1, $2, $3, 'PENDING')`
	_, err = tx.ExecContext(ctx, queryRes, orderID, productID, quantity)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresInventoryRepository) Release(ctx context.Context, orderID string, inventoryVersion int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get reservation details
	var productID uuid.UUID
	var quantity int
	var status string
	err = tx.QueryRowContext(ctx, "SELECT product_id, quantity, status FROM reservations WHERE order_id = $1", orderID).
		Scan(&productID, &quantity, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if status == "RELEASED" {
		return ErrIdempotent
	}
	if status != "PENDING" {
		return ErrInvalidState
	}

	// 2. Update inventory
	queryInv := `
		UPDATE inventory
		SET reserved = reserved - $1, version = version + 1, updated_at = NOW()
		WHERE product_id = $2 AND version = $3`
	res, err := tx.ExecContext(ctx, queryInv, quantity, productID, inventoryVersion)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrConflict
	}

	// 3. Update reservation status
	_, err = tx.ExecContext(ctx, "UPDATE reservations SET status = 'RELEASED', updated_at = NOW() WHERE order_id = $1", orderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresInventoryRepository) Commit(ctx context.Context, orderID string, inventoryVersion int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get reservation details
	var productID uuid.UUID
	var quantity int
	var status string
	err = tx.QueryRowContext(ctx, "SELECT product_id, quantity, status FROM reservations WHERE order_id = $1", orderID).
		Scan(&productID, &quantity, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if status == "COMMITTED" {
		return ErrIdempotent
	}
	if status != "PENDING" {
		return ErrInvalidState
	}

	// 2. Update inventory (absolute deduction)
	queryInv := `
		UPDATE inventory
		SET stock = stock - $1, reserved = reserved - $1, version = version + 1, updated_at = NOW()
		WHERE product_id = $2 AND version = $3`
	res, err := tx.ExecContext(ctx, queryInv, quantity, productID, inventoryVersion)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrConflict
	}

	// 3. Update reservation status
	_, err = tx.ExecContext(ctx, "UPDATE reservations SET status = 'COMMITTED', updated_at = NOW() WHERE order_id = $1", orderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresInventoryRepository) UpdateStock(ctx context.Context, productID string, totalStock int, currentVersion int) error {
	query := `
		UPDATE inventory
		SET stock = $1, version = version + 1, updated_at = NOW()
		WHERE product_id = $2 AND version = $3`

	res, err := r.db.ExecContext(ctx, query, totalStock, productID, currentVersion)
	if err != nil {
		return fmt.Errorf("could not update stock: %v", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrConflict
	}

	return nil
}
