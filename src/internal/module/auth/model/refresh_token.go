package model

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken представляет модель для работы с таблицей refresh_tokens
type RefreshToken struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"userId"`
	Token            string     `json:"token"`
	ExpiresAt        time.Time  `json:"expiresAt"`
	Revoked          bool       `json:"revoked"`
	CreatedAt        time.Time  `json:"createdAt"`
	CreatedByIP      string     `json:"createdByIp"`
	RevokedAt        *time.Time `json:"revokedAt,omitempty"`
	RevokedByIP      string     `json:"revokedByIp,omitempty"`
	ReplacedByToken  string     `json:"replacedByToken,omitempty"`
	DeviceIdentifier string     `json:"deviceIdentifier,omitempty"`
}

// IsExpired проверяет, истек ли срок действия токена
func (rt *RefreshToken) IsExpired() bool {
	return rt.ExpiresAt.Before(time.Now())
}

// IsActive проверяет, активен ли токен (не отозван и не истек)
func (rt *RefreshToken) IsActive() bool {
	return !rt.Revoked && !rt.IsExpired()
}

// Revoke отзывает токен
func (rt *RefreshToken) Revoke(ip string, replacedByToken string) {
	rt.Revoked = true
	now := time.Now()
	rt.RevokedAt = &now
	rt.RevokedByIP = ip
	rt.ReplacedByToken = replacedByToken
}
