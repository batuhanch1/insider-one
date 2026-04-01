package sms

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"insider-one/domain/notification"
	smsDomain "insider-one/domain/notification/sms"
	"insider-one/infrastructure/logging"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultBatchSize = 100

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) smsDomain.Repository {
	return &repository{pool}
}

func (r *repository) Save(ctx context.Context, sms smsDomain.Sms) error {
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
    idempotency_key
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
    NOW(),
    $9   -- idempotency_key
) ON CONFLICT DO NOTHING;`

	logging.DbQueryStart(ctx, query)
	result, err := r.pool.Exec(ctx, query, sms.PhoneNumber, sms.Sender, sms.Content, sms.Status, sms.Type, sms.ScheduledAt, sms.SentAt, sms.DeletedAt, sms.IdempotencyKey)
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

func buildFilter(status string, startDate, endDate *time.Time, page, pageSize int) (string, []any, int, int) {
	conditions := []string{"deleted_at IS NULL"}
	var args []any
	argIdx := 1

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	if startDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, startDate.Unix())
		argIdx++
	}

	if endDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, endDate.Unix())
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")
	offset := (page - 1) * pageSize

	return where, args, argIdx, offset
}

func (r *repository) List(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (smsDomain.SmsList, error) {
	where, args, argIdx, offset := buildFilter(status, startDate, endDate, page, pageSize)

	query := fmt.Sprintf(`
        SELECT
            id,
			phone_number,
			sender,
			content,
			status,
			type,
			scheduled_at,
			sent_at,
			deleted_at,
			created_at,
			idempotency_key
        FROM sms
        %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, where, argIdx, argIdx+1)

	args = append(args, pageSize, offset)

	logging.DbQueryStart(ctx, query)
	rows, err := r.pool.Query(ctx, query, args...)
	logging.DbQueryFinish(ctx)
	if err != nil {
		err = fmt.Errorf("list sms query error: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}
	defer rows.Close()

	var SmsList smsDomain.SmsList
	for rows.Next() {
		var sms smsDomain.Sms
		var scheduledAt, sentAt, deletedAt sql.NullTime

		err = rows.Scan(
			&sms.ID,
			&sms.PhoneNumber,
			&sms.Sender,
			&sms.Content,
			&sms.Status,
			&sms.Type,
			&scheduledAt,
			&sentAt,
			&deletedAt,
			&sms.CreatedAt,
			&sms.IdempotencyKey,
		)
		if err != nil {
			err = fmt.Errorf("list sms scan error: %w", err)
			logging.Error(ctx, err)
			return nil, err
		}

		if scheduledAt.Valid {
			sms.ScheduledAt = scheduledAt.Time.Unix()
		}
		if sentAt.Valid {
			sms.SentAt = sentAt.Time.Unix()
		}
		if deletedAt.Valid {
			sms.DeletedAt = deletedAt.Time.Unix()
		}

		SmsList = append(SmsList, sms)
	}

	if err = rows.Err(); err != nil {
		err = fmt.Errorf("list sms rows error: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}
	return SmsList, nil
}

func (r *repository) ListCount(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (int, error) {
	where, args, argIdx, offset := buildFilter(status, startDate, endDate, page, pageSize)

	query := fmt.Sprintf(`
        SELECT
            count(1) as totalCount
        FROM sms
        %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, where, argIdx, argIdx+1)

	args = append(args, pageSize, offset)

	var totalCount int
	logging.DbQueryStart(ctx, query)
	err := r.pool.QueryRow(ctx, query, args...).Scan(&totalCount)
	logging.DbQueryFinish(ctx)
	if err != nil {
		err = fmt.Errorf("listCount error: %w", err)
		logging.Error(ctx, err)
		return 0, err
	}
	return totalCount, nil
}

func (r *repository) GetByStatus(ctx context.Context, status string) ([]uint64, error) {
	var lastID uint64 = 0
	var totalIDs []uint64

	for {
		query := `
			SELECT id FROM sms
			WHERE status = $1
			  AND deleted_at IS NULL
			  AND id > $2
			ORDER BY id ASC
			LIMIT $3
		`

		logging.DbQueryStart(ctx, query)
		rows, err := r.pool.Query(ctx, query, status, lastID, defaultBatchSize)
		logging.DbQueryFinish(ctx)
		if err != nil {
			err = fmt.Errorf("getByStatus query: %w", err)
			logging.Error(ctx, err)
			return nil, err
		}

		ids := make([]uint64, 0, defaultBatchSize)
		for rows.Next() {
			var id uint64
			if err = rows.Scan(&id); err != nil {
				rows.Close()
				err = fmt.Errorf("getByStatus scan: %w", err)
				logging.Error(ctx, err)
				return nil, err
			}
			ids = append(ids, id)
		}

		if err = rows.Err(); err != nil {
			rows.Close()
			err = fmt.Errorf("getByStatus rows: %w", err)
			logging.Error(ctx, err)
			return nil, err
		}

		rows.Close()

		if len(ids) == 0 {
			break
		}

		totalIDs = append(totalIDs, ids...)
		lastID = ids[len(ids)-1]

		if len(ids) < defaultBatchSize {
			break
		}
	}

	return totalIDs, nil
}

func (r *repository) UpdateStatus(ctx context.Context, ids []uint64) error {
	query := `
        UPDATE sms
        SET status = $1
        WHERE id = ANY($2)
          AND deleted_at IS NULL
        RETURNING id
    `

	logging.DbQueryStart(ctx, query)
	result, err := r.pool.Exec(ctx, query, notification.Notification_Status_Pending, ids)
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

func (r *repository) GetStatusByID(ctx context.Context, ids []uint64) (smsDomain.SmsList, error) {
	var lastID uint64 = 0
	var totalSmsList smsDomain.SmsList

	for {
		query := `
			SELECT id, status FROM sms
			WHERE id = ANY($1)
			  AND deleted_at IS NULL
			  AND id > $2
			ORDER BY id ASC
			LIMIT $3
		`

		logging.DbQueryStart(ctx, query)
		rows, err := r.pool.Query(ctx, query, ids, lastID, defaultBatchSize)
		logging.DbQueryFinish(ctx)
		if err != nil {
			err = fmt.Errorf("getStatusByID query: %w", err)
			logging.Error(ctx, err)
			return nil, err
		}

		batch := make(smsDomain.SmsList, 0, defaultBatchSize)
		for rows.Next() {
			var sms smsDomain.Sms
			if err = rows.Scan(&sms.ID, &sms.Status); err != nil {
				rows.Close()
				err = fmt.Errorf("getStatusByID scan: %w", err)
				logging.Error(ctx, err)
				return nil, err
			}
			batch = append(batch, sms)
		}

		if err = rows.Err(); err != nil {
			rows.Close()
			err = fmt.Errorf("getStatusByID rows: %w", err)
			logging.Error(ctx, err)
			return nil, err
		}

		rows.Close()

		if len(batch) == 0 {
			break
		}

		totalSmsList = append(totalSmsList, batch...)
		lastID = batch[len(batch)-1].ID

		if len(batch) < defaultBatchSize {
			break
		}
	}

	return totalSmsList, nil
}

func (r *repository) Deliver(ctx context.Context, messageId string, idempotencyKey uint64) error {
	query := `
        UPDATE emails
        SET status = $1
        AND message_id = $3
        WHERE id = $2
          AND deleted_at IS NULL;`

	logging.DbQueryStart(ctx, query)
	result, err := r.pool.Exec(ctx, query, notification.Notification_Status_Delivered, idempotencyKey, messageId)
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
