package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/vishalss1/CartGO/services/product-service/internal/model"
)

type PostgresProductRepository struct {
	db *sql.DB
}

func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *model.Product) (*model.Product, error) {
	query := `
		INSERT INTO products (name, description, price, category, stock)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, p.Name, p.Description, p.Price, p.Category, p.Stock).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("could not create product: %v", err)
	}

	return p, nil
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string) (*model.Product, error) {
	query := `
		SELECT id, name, description, price, category, stock, created_at, updated_at
		FROM products
		WHERE id = $1`

	p := &model.Product{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category, &p.Stock, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("could not get product: %v", err)
	}

	return p, nil
}

func (r *PostgresProductRepository) List(ctx context.Context, filter *model.ProductFilter) ([]*model.Product, error) {
	var whereClauses []string
	var args []interface{}
	argCount := 1

	if filter.Category != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("category = $%d", argCount))
		args = append(args, filter.Category)
		argCount++
	}

	if filter.MinPrice != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("price >= $%d", argCount))
		args = append(args, *filter.MinPrice)
		argCount++
	}

	if filter.MaxPrice != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("price <= $%d", argCount))
		args = append(args, *filter.MaxPrice)
		argCount++
	}

	query := `
		SELECT id, name, description, price, category, stock, created_at, updated_at
		FROM products`

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not list products: %v", err)
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		p := &model.Product{}
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("could not scan product: %v", err)
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *model.Product) (*model.Product, error) {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, category = $4, stock = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query, p.Name, p.Description, p.Price, p.Category, p.Stock, p.ID).
		Scan(&p.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("could not update product: %v", err)
	}

	return p, nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("could not delete product: %v", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not get rows affected: %v", err)
	}

	if rows == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (r *PostgresProductRepository) UpdateStock(ctx context.Context, id string, delta int) error {
	// Delta can be positive (restock) or negative (sale)
	query := `
		UPDATE products
		SET stock = stock + $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND (stock + $1) >= 0`

	res, err := r.db.ExecContext(ctx, query, delta, id)
	if err != nil {
		return fmt.Errorf("could not update stock: %v", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not get rows affected: %v", err)
	}

	if rows == 0 {
		return fmt.Errorf("stock update failed: check product existence and sufficient stock")
	}

	return nil
}
