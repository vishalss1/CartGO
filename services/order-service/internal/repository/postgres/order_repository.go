package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/pkg/util"
	"github.com/vishalss1/CartGO/services/order-service/internal/model"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order *model.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if order.ID == uuid.Nil {
		order.ID = util.GenerateUUID()
	}

	// Insert order and return generated fields
	queryOrder := `
		INSERT INTO orders (id, user_id, total_amount, status, delivery_address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`
	err = tx.QueryRowContext(ctx, queryOrder,
		order.ID, order.UserID, order.TotalAmount, order.Status, order.DeliveryAddress).
		Scan(&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %v", err)
	}

	// Insert order items and return generated IDs
	queryItem := `
		INSERT INTO order_items (id, order_id, product_id, quantity, price_per_unit)
		VALUES ($1, $2, $3, $4, $5)`
	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		order.Items[i].ID = util.GenerateUUID()
		_, err = tx.ExecContext(ctx, queryItem,
			order.Items[i].ID, order.Items[i].OrderID, order.Items[i].ProductID, order.Items[i].Quantity, order.Items[i].PricePerUnit)
		if err != nil {
			return fmt.Errorf("failed to insert order item at index %d: %v", i, err)
		}
	}

	return tx.Commit()
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	queryOrder := `
		SELECT id, user_id, total_amount, status, delivery_address, created_at, updated_at
		FROM orders
		WHERE id = $1`

	order := &model.Order{}
	err := r.db.QueryRowContext(ctx, queryOrder, id).Scan(
		&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.DeliveryAddress, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Get items
	queryItems := `
		SELECT id, order_id, product_id, quantity, price_per_unit
		FROM order_items
		WHERE order_id = $1`
	rows, err := r.db.QueryContext(ctx, queryItems, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := model.OrderItem{}
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.PricePerUnit)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *PostgresOrderRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Order, error) {
	query := `
		SELECT id, user_id, total_amount, status, delivery_address, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		order := &model.Order{}
		err := rows.Scan(&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.DeliveryAddress, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
