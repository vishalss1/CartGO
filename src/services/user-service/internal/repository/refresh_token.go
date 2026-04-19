package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/pkg/util"
	"github.com/vishalss1/CartGO/services/user-service/internal/model"
)

type RefreshTokenRepository interface {
	StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
}

type PostgresRefreshTokenRepository struct {
	db *sql.DB
}

func NewPostgresRefreshTokenRepository(db *sql.DB) RefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

func (r *PostgresRefreshTokenRepository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	id := util.GenerateUUID()
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, id, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %v", err)
	}

	return nil
}

func (r *PostgresRefreshTokenRepository) GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1`

	var rt model.RefreshToken
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %v", err)
	}

	return &rt, nil
}

func (r *PostgresRefreshTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %v", err)
	}
	return nil
}

func (r *PostgresRefreshTokenRepository) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user refresh tokens: %v", err)
	}
	return nil
}
