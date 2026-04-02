package sms

import (
	"context"
	"errors"
	"fmt"
	"insider-one/domain/notification"
	smsDomain "insider-one/domain/notification/sms"
	"insider-one/infrastructure/logging"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type commandRepository struct {
	pool *pgxpool.Pool
}

func NewCommandRepository(pool *pgxpool.Pool) smsDomain.CommandRepository {
	return &commandRepository{pool}
}
func (r *commandRepository) Save(ctx context.Context, sms smsDomain.Sms) error {
	query := `
INSERT INTO sms (
    phone_number,
    sender,
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
    $1,  -- phone_number
    $2,  -- sender
    $3,  -- content
    $4,  -- status
    $5,  -- type
    $6,  -- scheduled_at (nullable)
    $7,  -- sent_at (nullable)
    $8,  -- deleted_at (nullable)
    $10,
    $9,   -- idempotency_key
    $11
) ON CONFLICT DO NOTHING;`

	logging.DbQueryStart(ctx, query)
	result, err := r.pool.Exec(ctx, query, sms.PhoneNumber, sms.Sender, sms.Content, sms.Status, sms.Type, sms.ScheduledAt, sms.SentAt, sms.DeletedAt, sms.IdempotencyKey, time.Now().Unix(), sms.Priority)
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
func (r *commandRepository) Deliver(ctx context.Context, messageId string, idempotencyKey uint64) error {
	query := `
        UPDATE sms
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
func (r *commandRepository) Cancel(ctx context.Context, id uint64) error {
	query := `
        UPDATE sms
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
		logging.Error(ctx, errors.New("no row updated"))
	}

	return nil
}
