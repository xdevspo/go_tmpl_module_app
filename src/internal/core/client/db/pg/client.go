package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"github.com/xdevspo/go_tmpl_module_app/internal/core/client/db"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/logger"
)

type pgClient struct {
	masterDBC db.DB
}

func New(ctx context.Context, dsn string, logger logger.Logger) (db.Client, error) {
	dbc, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, errors.Errorf("failed to connect to db: %v", err)
	}

	return &pgClient{
		masterDBC: NewDB(dbc, logger),
	}, nil
}

func (c *pgClient) DB() db.DB {
	return c.masterDBC
}

func (c *pgClient) Close() error {
	if c.masterDBC != nil {
		c.masterDBC.Close()
	}

	return nil
}
