package service

import (
	"context"

	"github.com/google/uuid"
	authModel "github.com/xdevspo/go_tmpl_module_app/internal/module/auth/model"
	userModel "github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Register creates a new user account and returns authentication response
	Register(ctx context.Context, req *userModel.CreateUserRequest) (*authModel.AuthResponse, error)

	// Login authenticates a user and returns authentication response
	Login(ctx context.Context, email, password string) (*authModel.AuthResponse, error)

	// RefreshToken refreshes an access token using a refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*authModel.AuthResponse, error)

	// RevokeToken отзывает указанный refresh токен
	RevokeToken(ctx context.Context, tokenStr string, ipAddress string) error

	// RevokeAllUserTokens отзывает все токены пользователя
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, ipAddress string) error

	// GetRefreshTokens возвращает все активные токены пользователя
	GetRefreshTokens(ctx context.Context, userID uuid.UUID) ([]authModel.RefreshToken, error)
}
