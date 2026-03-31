package postgresql

import (
	"context"
	"fmt"
	"insider-one/infrastructure/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(dbConfig config.DBConfig, ctx context.Context) (*pgxpool.Pool, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s:%v/%s?sslmode=disable", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name)
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("config parse error: %w", err)
	}

	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pool error: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping error: %w", err)
	}

	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
CREATE TABLE IF NOT EXISTS emails (
	id                BIGINT          PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	"to"              VARCHAR(320)    NOT NULL,
	"from"            VARCHAR(320)    NOT NULL,
	subject           VARCHAR(998)    NOT NULL,
	content           TEXT            NOT NULL,
	status            VARCHAR(20)     NOT NULL DEFAULT 'pending',
	notification_type VARCHAR(50)     NOT NULL,
	scheduled_at      TIMESTAMPTZ,
	sent_at           TIMESTAMPTZ,
	deleted_at        TIMESTAMPTZ,
	created_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

	CONSTRAINT chk_to_format CHECK ("to" LIKE '%@%'),
	CONSTRAINT chk_from_format CHECK ("from" LIKE '%@%')
);

CREATE INDEX IF NOT EXISTS idx_email_status_created ON email_notifications(status, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_email_scheduled_at ON email_notifications(scheduled_at) WHERE status = 'pending' AND scheduled_at IS NOT NULL;
`
	if _, err := pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("migration error: %w", err)
	}
	return nil
}
