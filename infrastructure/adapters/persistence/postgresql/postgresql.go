package postgresql

import (
	"context"
	"fmt"
	"insider-one/infrastructure/config"
	"insider-one/infrastructure/logging"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(dbConfig config.DBConfig, ctx context.Context) (*pgxpool.Pool, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s:%v/%s?sslmode=disable", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name)
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		err = fmt.Errorf("config parse error: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		err = fmt.Errorf("pool error: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		err = fmt.Errorf("ping error: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
CREATE TABLE IF NOT EXISTS emails (
	id              BIGINT       PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	"to"            VARCHAR(320) NOT NULL,
	"from"          VARCHAR(320) NOT NULL,
	subject         VARCHAR(998) NOT NULL,
	content         TEXT         NOT NULL,
	status          VARCHAR(20)  NOT NULL,
	type            VARCHAR(50)  NOT NULL,
	priority        VARCHAR(10)  NOT NULL,
	idempotency_key NUMERIC(20,0)       UNIQUE,
	message_id      VARCHAR(255),
	scheduled_at    BIGINT,
	sent_at         BIGINT,
	deleted_at      BIGINT,
	created_at      BIGINT  NOT NULL);

CREATE INDEX IF NOT EXISTS idx_emails_status_created_at
  ON emails(status, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_emails_status_id
  ON emails(status, id ASC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_emails_scheduled_at
  ON emails(scheduled_at ASC)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS pushes (
	id              BIGINT       PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	phone_number    VARCHAR(20)  NOT NULL,
	sender          VARCHAR(100) NOT NULL,
	content         TEXT         NOT NULL,
	status          VARCHAR(20)  NOT NULL,
	type            VARCHAR(50)  NOT NULL,
	priority        VARCHAR(10)  NOT NULL,
	idempotency_key NUMERIC(20,0)       UNIQUE,
	message_id      VARCHAR(255),
	scheduled_at    BIGINT,
	sent_at         BIGINT,
	deleted_at      BIGINT,
	created_at      BIGINT  NOT NULL);

CREATE INDEX IF NOT EXISTS idx_pushes_status_created_at
  ON pushes(status, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_pushes_status_id
  ON pushes(status, id ASC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_pushes_scheduled_at
  ON pushes(scheduled_at ASC)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS sms (
	id              BIGINT      PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	phone_number    VARCHAR(20) NOT NULL,
	sender          VARCHAR(11) NOT NULL,
	content         TEXT        NOT NULL,
	status          VARCHAR(20) NOT NULL,
	type            VARCHAR(50) NOT NULL,
	priority        VARCHAR(10) NOT NULL,
	idempotency_key NUMERIC(20,0)      UNIQUE,
	message_id      VARCHAR(255),
	scheduled_at    BIGINT,
	sent_at         BIGINT,
	deleted_at      BIGINT,
	created_at      BIGINT NOT NULL);

CREATE INDEX IF NOT EXISTS idx_sms_status_created_at
  ON sms(status, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sms_status_id
  ON sms(status, id ASC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sms_scheduled_at
  ON sms(scheduled_at ASC)
  WHERE deleted_at IS NULL;
`
	if _, err := pool.Exec(ctx, query); err != nil {
		err = fmt.Errorf("migration error: %w", err)
		logging.Error(ctx, err)
		return err
	}
	return nil
}
