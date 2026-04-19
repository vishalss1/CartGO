package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/user-service/internal/auth"
	"github.com/vishalss1/CartGO/services/user-service/internal/config"
	"github.com/vishalss1/CartGO/services/user-service/internal/model"
	"github.com/vishalss1/CartGO/services/user-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

type UserService interface {
	Signup(ctx context.Context, req model.SignupRequest) (*model.AuthResponse, error)
	Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*model.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	ListAllUsers(ctx context.Context) ([]*model.User, error)
	ChangeUserRole(ctx context.Context, userID uuid.UUID, newRole string) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) (*model.User, error)
}

type UserServiceImpl struct {
	userRepo  repository.UserRepository
	tokenRepo repository.RefreshTokenRepository
	config    *config.Config
}

func NewUserService(userRepo repository.UserRepository, tokenRepo repository.RefreshTokenRepository, cfg *config.Config) UserService {
	return &UserServiceImpl{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		config:    cfg,
	}
}

func (s *UserServiceImpl) Signup(ctx context.Context, req model.SignupRequest) (*model.AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
	}

	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	return s.createAuthResponse(ctx, user)
}

func (s *UserServiceImpl) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	// Get user
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	return s.createAuthResponse(ctx, user)
}

func (s *UserServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*model.AuthResponse, error) {
	// Validate token in DB
	rt, err := s.tokenRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	if rt == nil {
		return nil, ErrInvalidToken
	}

	// Validate token signature and expiry
	_, err = auth.ValidateToken(refreshToken, s.config.JWTPublicKeys)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	// Delete old token (rotation)
	_ = s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)

	// Generate new tokens
	return s.createAuthResponse(ctx, user)
}

func (s *UserServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	return s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
}

func (s *UserServiceImpl) ListAllUsers(ctx context.Context) ([]*model.User, error) {
	return s.userRepo.GetAllUsers(ctx)
}

func (s *UserServiceImpl) ChangeUserRole(ctx context.Context, userID uuid.UUID, newRole string) error {
	return s.userRepo.UpdateUserRole(ctx, userID, newRole)
}

func (s *UserServiceImpl) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	return s.userRepo.GetUserByID(ctx, userID)
}

func (s *UserServiceImpl) UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" && req.Email != user.Email {
		existingUser, checkErr := s.userRepo.GetUserByEmail(ctx, req.Email)
		if checkErr != nil {
			return nil, checkErr
		}
		if existingUser != nil {
			return nil, ErrUserAlreadyExists
		}
		user.Email = req.Email
	}

	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserServiceImpl) createAuthResponse(ctx context.Context, user *model.User) (*model.AuthResponse, error) {
	accessToken, err := auth.GenerateAccessToken(user.ID.String(), user.Role, s.config.JWTPrivateKey, s.config.JWTPrivateKeyID, s.config.AccessTokenExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpiry, err := auth.GenerateRefreshToken(user.ID.String(), s.config.JWTPrivateKey, s.config.JWTPrivateKeyID, s.config.RefreshTokenExpiry)
	if err != nil {
		return nil, err
	}

	// Store refresh token in DB
	err = s.tokenRepo.StoreRefreshToken(ctx, user.ID, refreshToken, refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}
