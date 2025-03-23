package model

type TokenClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
}
