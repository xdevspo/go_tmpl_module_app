package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// UserClaims расширяет стандартные JWT claims
type UserClaims struct {
	jwt.RegisteredClaims
	UserID      string   `json:"user_id"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// Manager предоставляет методы для работы с JWT токенами
type Manager struct {
	secret            string
	expirationMinutes int
}

// NewManager создает новый экземпляр JWT Manager
func NewManager(secret string, expirationMinutes int) *Manager {
	return &Manager{
		secret:            secret,
		expirationMinutes: expirationMinutes,
	}
}

// GenerateToken создает JWT токен для пользователя с указанными ролями
func (m *Manager) GenerateToken(userID string, roles []string, permissions []string) (string, error) {
	// Время жизни токена
	expiresAt := time.Now().Add(time.Duration(m.expirationMinutes) * time.Minute)

	// Создание claims
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
		UserID:      userID,
		Roles:       roles,
		Permissions: permissions,
	}

	// Создание токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписание токена
	tokenString, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", fmt.Errorf("ошибка подписания токена: %w", err)
	}

	return tokenString, nil
}

// ValidateToken проверяет токен и возвращает его claims
func (m *Manager) ValidateToken(tokenString string) (*UserClaims, error) {
	// Парсинг токена с проверкой подписи
	token, err := jwt.ParseWithClaims(
		tokenString,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Проверяем, что используется правильный алгоритм подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
			}
			return []byte(m.secret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("ошибка валидации токена: %w", err)
	}

	// Проверяем, что токен действителен
	if !token.Valid {
		return nil, errors.New("недействительный токен")
	}

	// Получаем claims из токена
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, errors.New("невозможно получить claims из токена")
	}

	return claims, nil
}

// ExpirationMinutes возвращает время жизни токена в минутах
func (m *Manager) ExpirationMinutes() int {
	return m.expirationMinutes
}

// Secret возвращает секретный ключ для подписи JWT
func (m *Manager) Secret() string {
	return m.secret
}

// HasRole проверяет наличие указанной роли в claims
func (claims *UserClaims) HasRole(role string) bool {
	for _, r := range claims.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole проверяет наличие хотя бы одной из указанных ролей в claims
func (claims *UserClaims) HasAnyRole(roles ...string) bool {
	for _, requiredRole := range roles {
		if claims.HasRole(requiredRole) {
			return true
		}
	}
	return false
}
