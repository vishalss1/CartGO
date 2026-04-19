package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/user-service/internal/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	GetAllUsers(ctx context.Context) ([]*model.User, error)
	UpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error
}
