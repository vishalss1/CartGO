package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/vishalss1/CartGO/services/user-service/internal/model"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		now,
		now,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %v", err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %v", err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, password_hash = $3, role = $4, updated_at = $5
		WHERE id = $6`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		now,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	user.UpdatedAt = now
	return nil
}

func (r *PostgresUserRepository) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	query := `SELECT id, username, email, role, created_at, updated_at FROM users`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %v", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

func (r *PostgresUserRepository) UpdateUserRole(ctx context.Context, userID string, role string) error {
	query := `UPDATE users SET role = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, role, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user role: %v", err)
	}
	return nil
}
