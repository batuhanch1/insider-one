package email

import (
	"context"
	"errors"
	"fmt"
	"insider-one/domain/notification"
	emailDomain "insider-one/domain/notification/email"
	"insider-one/infrastructure/logging"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type commandRepository struct {
	pool *pgxpool.Pool
}

func NewCommandRepository(pool *pgxpool.Pool) emailDomain.CommandRepository {
	return &commandRepository{pool}
}

func (r *commandRepository) Save(ctx context.Context, email emailDomain.Email) error {
	query := `
INSERT INTO emails (
    "to",
    "from",
    subject,
    content,
    status,
    type,
    scheduled_at,
    sent_at,
    deleted_at,
    created_at,
	idempotency_key,
	priority
)
VALUES (
    $1,  -- to
    $2,  -- from
    $3,  -- subject
    $4,  -- content
    $5,  -- status
    $6,  -- type
    $7,  -- scheduled_at (nullable)
    $8,  -- sent_at (nullable)
    $9,  -- deleted_at (nullable)
    $11,
	$10,  -- idempotency_key
	$12
) ON CONFLICT DO NOTHING;`

	logging.DbQueryStart(ctx, query)
	result, err := r.pool.Exec(ctx, query, email.To, email.From, email.Subject, email.Content, email.Status, email.Type, email.ScheduledAt, nil, nil, email.IdempotencyKey, time.Now().Unix(), email.Priority)
	logging.DbQueryFinish(ctx)
	if err != nil {
		err = fmt.Errorf("save error: %w", err)
		logging.Error(ctx, err)
		return err
	}
	if result.RowsAffected() == 0 {
		logging.Error(ctx, errors.New("rowsAffected is zero. save error"))
	}
	return nil
}

func (r *commandRepository) Cancel(ctx context.Context, id uint64) error {
	query := `
        UPDATE emails
        SET status = $1
        WHERE id = $2
          AND deleted_at IS NULL
        RETURNING id
    `

	logging.DbQueryStart(ctx, query)
	result, err := r.pool.Exec(ctx, query, notification.Notification_Status_Cancel, id)
	logging.DbQueryFinish(ctx)
	if err != nil {
		err = fmt.Errorf("update status query: %w", err)
		logging.Error(ctx, err)
		return err
	}
	if result.RowsAffected() == 0 {
		err = errors.New("no row updated")
		logging.Error(ctx, err)
	}

	return nil
}

func (r *commandRepository) Deliver(ctx context.Context, messageId string, idempotencyKey uint64) error {
	query := `
        UPDATE emails
        SET status = $1, message_id = $3, sent_at = $4
        WHERE idempotency_key = $2
          AND deleted_at IS NULL;`

	logging.DbQueryStart(ctx, query)
	result, err := r.pool.Exec(ctx, query, notification.Notification_Status_Delivered, idempotencyKey, messageId, time.Now().Unix())
	logging.DbQueryFinish(ctx)
	if err != nil {
		err = fmt.Errorf("deliver query: %w", err)
		logging.Error(ctx, err)
		return err
	}
	if result.RowsAffected() == 0 {
		logging.Error(ctx, errors.New("no row updated"))
	}

	return nil
}
