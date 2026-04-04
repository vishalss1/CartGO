package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
}

type UserServiceImpl struct {
	userRepo   repository.UserRepository
	tokenRepo  repository.RefreshTokenRepository
	config     *config.Config
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
	if rt == nil || time.Now().After(rt.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	// Optional: Delete old token for rotation
	_ = s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)

	// Generate new tokens
	return s.createAuthResponse(ctx, user)
}

func (s *UserServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	return s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
}

func (s *UserServiceImpl) createAuthResponse(ctx context.Context, user *model.User) (*model.AuthResponse, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpiry, err := s.generateRefreshToken(user)
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

func (s *UserServiceImpl) generateAccessToken(user *model.User) (string, error) {
	expiry, err := time.ParseDuration(s.config.AccessTokenExpiry)
	if err != nil {
		expiry = time.Minute * 15
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.AccessTokenSecret))
}

func (s *UserServiceImpl) generateRefreshToken(user *model.User) (string, time.Time, error) {
	expiry, err := time.ParseDuration(s.config.RefreshTokenExpiry)
	if err != nil {
		expiry = time.Hour * 24 * 7
	}

	expiresAt := time.Now().Add(expiry)
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.config.RefreshTokenSecret))
	return tokenStr, expiresAt, err
}
