package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/pkg/util"
	"github.com/vishalss1/CartGO/services/product-service/internal/model"
)

type PostgresProductRepository struct {
	db *sql.DB
}

func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *model.Product) (*model.Product, error) {
	if p.ID == uuid.Nil {
		p.ID = util.GenerateUUID()
	}

	query := `
		INSERT INTO products (id, name, description, price, category)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, p.ID, p.Name, p.Description, p.Price, p.Category).
		Scan(&p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("could not create product: %v", err)
	}

	return p, nil
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	query := `
		SELECT id, name, description, price, category, created_at, updated_at
		FROM products
		WHERE id = $1`

	p := &model.Product{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, err // Let service handle error translation
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
		SELECT id, name, description, price, category, created_at, updated_at
		FROM products`

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY created_at DESC"

	// Add Pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
		argCount++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not list products: %v", err)
	}
	defer rows.Close()

	products := []*model.Product{}
	for rows.Next() {
		p := &model.Product{}
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category, &p.CreatedAt, &p.UpdatedAt)
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
		SET name = $1, description = $2, price = $3, category = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query, p.Name, p.Description, p.Price, p.Category, p.ID).
		Scan(&p.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
