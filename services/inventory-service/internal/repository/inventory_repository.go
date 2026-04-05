package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vishalss1/CartGO/services/inventory-service/internal/model"
)

var (
	ErrConflict = errors.New("inventory record was modified by another process")
	ErrNotFound = errors.New("inventory record not found")
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
		ON CONFLICT (product_id) DO ACTION
		RETURNING version, created_at, updated_at`

	// This is a simplified upsert. For a real system, we might want to handle existing records differently.
	err := r.db.QueryRowContext(ctx, query, inv.ProductID, inv.Stock, inv.Reserved).
		Scan(&inv.Version, &inv.CreatedAt, &inv.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("could not upsert inventory: %v", err)
	}

	return inv, nil
}

func (r *PostgresInventoryRepository) Reserve(ctx context.Context, productID string, quantity int, currentVersion int) error {
	query := `
		UPDATE inventory
		SET reserved = reserved + $1, version = version + 1, updated_at = NOW()
		WHERE product_id = $2 AND (stock - reserved) >= $1 AND version = $3`

	res, err := r.db.ExecContext(ctx, query, quantity, productID, currentVersion)
	if err != nil {
		return fmt.Errorf("could not reserve stock: %v", err)
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

func (r *PostgresInventoryRepository) Release(ctx context.Context, productID string, quantity int, currentVersion int) error {
	query := `
		UPDATE inventory
		SET reserved = reserved - $1, version = version + 1, updated_at = NOW()
		WHERE product_id = $2 AND reserved >= $1 AND version = $3`

	res, err := r.db.ExecContext(ctx, query, quantity, productID, currentVersion)
	if err != nil {
		return fmt.Errorf("could not release stock: %v", err)
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

func (r *PostgresInventoryRepository) Commit(ctx context.Context, productID string, quantity int, currentVersion int) error {
	query := `
		UPDATE inventory
		SET stock = stock - $1, reserved = reserved - $1, version = version + 1, updated_at = NOW()
		WHERE product_id = $2 AND stock >= $1 AND reserved >= $1 AND version = $3`

	res, err := r.db.ExecContext(ctx, query, quantity, productID, currentVersion)
	if err != nil {
		return fmt.Errorf("could not commit stock: %v", err)
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
