package config

import (
	"strconv"
)

const (
	secretKey                = "JWT_SECRET_KEY"
	accessTokenExpiryMinutes = "ACCESS_TOKEN_EXPIRY_MINUTES"
	refreshTokenExpiryHours  = "REFRESH_TOKEN_EXPIRY_HOURS"
)

type JWTConfig interface {
	SecretKey() string
	AccessTokenExpiryMinutes() int
	RefreshTokenExpiryHours() int
}

type jwtConfig struct {
	secretKey                string
	accessTokenExpiryMinutes int
	refreshTokenExpiryHours  int
}

func NewJWTConfig() (JWTConfig, error) {
	secretKey := getEnv(secretKey, "sa!5da#54d3@4")
	accessTokenExpiryMinutes, _ := strconv.Atoi(getEnv(accessTokenExpiryMinutes, "60"))
	refreshTokenExpiryHours, _ := strconv.Atoi(getEnv(refreshTokenExpiryHours, "24"))

	return &jwtConfig{
		secretKey:                secretKey,
		accessTokenExpiryMinutes: accessTokenExpiryMinutes,
		refreshTokenExpiryHours:  refreshTokenExpiryHours,
	}, nil
}

func (cfg *jwtConfig) SecretKey() string {
	return cfg.secretKey
}

func (cfg *jwtConfig) AccessTokenExpiryMinutes() int {
	return cfg.accessTokenExpiryMinutes
}

func (cfg *jwtConfig) RefreshTokenExpiryHours() int {
	return cfg.refreshTokenExpiryHours
}
