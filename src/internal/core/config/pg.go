package config

import (
	"errors"
)

const (
	dsnEnvName = "POSTGRES_DSN"
)

type PGConfig interface {
	DSN() string
}

type pgConfig struct {
	dsn string
}

func NewPGConfig() (PGConfig, error) {
	dsn := getEnv(dsnEnvName, "localhost")
	if len(dsn) == 0 {
		return nil, errors.New("POSTGRES_DSN is not set")
	}

	return &pgConfig{
		dsn: dsn,
	}, nil
}

func (cfg *pgConfig) DSN() string {
	return cfg.dsn
}
