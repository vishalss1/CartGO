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

func (r *PostgresInventoryRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*model.Inventory, error) {
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

func (r *PostgresInventoryRepository) GetReservations(ctx context.Context, orderID uuid.UUID) ([]*model.Reservation, error) {
	query := `
		SELECT order_id, product_id, quantity, status, created_at, updated_at
		FROM reservations
		WHERE order_id = $1`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("could not query reservations: %v", err)
	}
	defer rows.Close()

	var reservations []*model.Reservation
	for rows.Next() {
		res := &model.Reservation{}
		err := rows.Scan(&res.OrderID, &res.ProductID, &res.Quantity, &res.Status, &res.CreatedAt, &res.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("could not scan reservation: %v", err)
		}
		reservations = append(reservations, res)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(reservations) == 0 {
		return nil, ErrNotFound
	}

	return reservations, nil
}

func (r *PostgresInventoryRepository) Reserve(ctx context.Context, productID uuid.UUID, orderID uuid.UUID, quantity int, currentVersion int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check for existing reservation for this SPECIFIC product in this order (Idempotency)
	var existingStatus string
	err = tx.QueryRowContext(ctx, "SELECT status FROM reservations WHERE order_id = $1 AND product_id = $2", orderID, productID).Scan(&existingStatus)
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

func (r *PostgresInventoryRepository) Release(ctx context.Context, orderID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get ALL pending reservations for this order
	query := "SELECT product_id, quantity, status FROM reservations WHERE order_id = $1 FOR UPDATE"
	rows, err := tx.QueryContext(ctx, query, orderID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type item struct {
		productID uuid.UUID
		quantity  int
		status    string
	}
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.productID, &it.quantity, &it.status); err != nil {
			return err
		}
		items = append(items, it)
	}

	if len(items) == 0 {
		return ErrNotFound
	}

	// 2. Process each item
	for _, it := range items {
		if it.status == "RELEASED" {
			continue
		}
		if it.status != "PENDING" {
			return fmt.Errorf("%w: product %s is in state %s", ErrInvalidState, it.productID, it.status)
		}

		// Update inventory (we don't check version here to keep it simple across multiple items, 
		// but we could fetch versions first if needed. Absolute subtraction of reserved is safe 
		// as long as we're in a transaction and used FOR UPDATE above)
		queryInv := `
			UPDATE inventory
			SET reserved = reserved - $1, version = version + 1, updated_at = NOW()
			WHERE product_id = $2`
		_, err := tx.ExecContext(ctx, queryInv, it.quantity, it.productID)
		if err != nil {
			return err
		}
	}

	// 3. Update all reservation statuses for the order
	_, err = tx.ExecContext(ctx, "UPDATE reservations SET status = 'RELEASED', updated_at = NOW() WHERE order_id = $1 AND status = 'PENDING'", orderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresInventoryRepository) Commit(ctx context.Context, orderID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get ALL pending reservations for this order
	query := "SELECT product_id, quantity, status FROM reservations WHERE order_id = $1 FOR UPDATE"
	rows, err := tx.QueryContext(ctx, query, orderID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type item struct {
		productID uuid.UUID
		quantity  int
		status    string
	}
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.productID, &it.quantity, &it.status); err != nil {
			return err
		}
		items = append(items, it)
	}

	if len(items) == 0 {
		return ErrNotFound
	}

	// 2. Process each item
	for _, it := range items {
		if it.status == "COMMITTED" {
			continue
		}
		if it.status != "PENDING" {
			return fmt.Errorf("%w: product %s is in state %s", ErrInvalidState, it.productID, it.status)
		}

		// Update inventory (deduct from both stock and reserved)
		queryInv := `
			UPDATE inventory
			SET stock = stock - $1, reserved = reserved - $1, version = version + 1, updated_at = NOW()
			WHERE product_id = $2`
		_, err := tx.ExecContext(ctx, queryInv, it.quantity, it.productID)
		if err != nil {
			return err
		}
	}

	// 3. Update all reservation statuses for the order
	_, err = tx.ExecContext(ctx, "UPDATE reservations SET status = 'COMMITTED', updated_at = NOW() WHERE order_id = $1 AND status = 'PENDING'", orderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresInventoryRepository) UpdateStock(ctx context.Context, productID uuid.UUID, totalStock int, currentVersion int) error {
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

func (r *PostgresInventoryRepository) AdjustStock(ctx context.Context, productID uuid.UUID, delta int) error {
	query := `
		UPDATE inventory
		SET stock = stock + $1, version = version + 1, updated_at = NOW()
		WHERE product_id = $2`

	res, err := r.db.ExecContext(ctx, query, delta, productID)
	if err != nil {
		return fmt.Errorf("could not adjust stock: %v", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
