package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/auth/model"
)

// RefreshTokenRepository определяет интерфейс для операций с refresh токенами
type RefreshTokenRepository interface {
	// Create создает новый refresh токен в базе данных
	Create(ctx context.Context, token *model.RefreshToken) error

	// GetByToken находит токен по его значению
	GetByToken(ctx context.Context, token string) (*model.RefreshToken, error)

	// GetByUserID находит все активные токены пользователя
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]model.RefreshToken, error)

	// Update обновляет информацию о токене
	Update(ctx context.Context, token *model.RefreshToken) error

	// RevokeAll отзывает все активные токены пользователя
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, ipAddress string) error

	// DeleteExpired удаляет все истекшие токены
	DeleteExpired(ctx context.Context) error

	// CountActiveByUserID возвращает количество активных токенов пользователя
	CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// RevokeOldestIfLimitExceeded отзывает самые старые токены, если превышен лимит
	RevokeOldestIfLimitExceeded(ctx context.Context, userID uuid.UUID, limit int, ipAddress string) error

	// GetByDeviceIdentifier находит токен по идентификатору устройства
	GetByDeviceIdentifier(ctx context.Context, userID uuid.UUID, deviceID string) (*model.RefreshToken, error)

	// RevokeByDeviceIdentifier отзывает токен для конкретного устройства
	RevokeByDeviceIdentifier(ctx context.Context, userID uuid.UUID, deviceID string, ipAddress string) error
}
