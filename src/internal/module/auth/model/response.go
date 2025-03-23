package model

import (
	"time"

	"github.com/google/uuid"
)

type AuthResponse struct {
	Token           string    `json:"accessToken"`
	RefreshToken    string    `json:"refreshToken,omitempty"`
	TokenType       string    `json:"token_type"`
	ExpiresIn       int64     `json:"expires_in"`
	AccessExpiresAt time.Time `json:"accessExpiresAt,omitempty"`
}

type RegisterResponse struct {
	ID uuid.UUID `json:"id"`
}
