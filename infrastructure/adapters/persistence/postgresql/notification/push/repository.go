package push

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"insider-one/domain/notification"
	PushDomain "insider-one/domain/notification/push"
	pushDomain "insider-one/domain/notification/push"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultBatchSize = 100

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) pushDomain.Repository {
	return &repository{pool}
}

func (r *repository) Save(ctx context.Context, push pushDomain.Push) error {
	query := `
INSERT INTO pushes (
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

	result, err := r.pool.Exec(ctx, query, push.PhoneNumber, push.Sender, push.Content, push.Status, push.Type, push.ScheduledAt, push.SentAt, push.DeletedAt, push.IdempotencyKey)
	if err != nil {
		return fmt.Errorf("save error: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("rowsAffected is zero. save error")
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

func (r *repository) List(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (pushDomain.Pushes, error) {
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
        FROM pushes
        %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, where, argIdx, argIdx+1)

	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list push query error: %w", err)
	}
	defer rows.Close()

	var pushes pushDomain.Pushes
	for rows.Next() {
		var push PushDomain.Push
		var scheduledAt, sentAt, deletedAt sql.NullTime

		err = rows.Scan(
			&push.ID,
			&push.PhoneNumber,
			&push.Sender,
			&push.Content,
			&push.Status,
			&push.Type,
			&scheduledAt,
			&sentAt,
			&deletedAt,
			&push.CreatedAt,
			&push.IdempotencyKey,
		)
		if err != nil {
			return nil, fmt.Errorf("list push scan error: %w", err)
		}

		if scheduledAt.Valid {
			push.ScheduledAt = scheduledAt.Time.Unix()
		}
		if sentAt.Valid {
			push.SentAt = sentAt.Time.Unix()
		}
		if deletedAt.Valid {
			push.DeletedAt = deletedAt.Time.Unix()
		}

		pushes = append(pushes, push)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list push rows error: %w", err)
	}
	return pushes, nil
}

func (r *repository) ListCount(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (int, error) {
	where, args, argIdx, offset := buildFilter(status, startDate, endDate, page, pageSize)

	query := fmt.Sprintf(`
        SELECT
            count(1) as totalCount
        FROM pushes
        %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, where, argIdx, argIdx+1)

	args = append(args, pageSize, offset)

	var totalCount int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("listCount error: %w", err)
	}
	return totalCount, nil
}

func (r *repository) GetByStatus(ctx context.Context, status string) ([]uint64, error) {
	var lastID uint64 = 0
	var totalIDs []uint64

	for {
		query := `
			SELECT id FROM pushes
			WHERE status = $1
			  AND deleted_at IS NULL
			  AND id > $2
			ORDER BY id ASC
			LIMIT $3
		`

		rows, err := r.pool.Query(ctx, query, status, lastID, defaultBatchSize)
		if err != nil {
			return nil, fmt.Errorf("getByStatus query: %w", err)
		}

		ids := make([]uint64, 0, defaultBatchSize)
		for rows.Next() {
			var id uint64
			if err = rows.Scan(&id); err != nil {
				rows.Close()
				return nil, fmt.Errorf("getByStatus scan: %w", err)
			}
			ids = append(ids, id)
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("getByStatus rows: %w", err)
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
        UPDATE pushes
        SET status = $1
        WHERE id = ANY($2)
          AND deleted_at IS NULL
        RETURNING id
    `

	result, err := r.pool.Exec(ctx, query, notification.Notification_Status_Pending, ids)
	if err != nil {
		return fmt.Errorf("update status query: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("no row updated")
	}

	return nil
}

func (r *repository) GetStatusByID(ctx context.Context, ids []uint64) (pushDomain.Pushes, error) {
	var lastID uint64 = 0
	var totalpushes pushDomain.Pushes

	for {
		query := `
			SELECT id, status FROM pushes
			WHERE id = ANY($1)
			  AND deleted_at IS NULL
			  AND id > $2
			ORDER BY id ASC
			LIMIT $3
		`

		rows, err := r.pool.Query(ctx, query, ids, lastID, defaultBatchSize)
		if err != nil {
			return nil, fmt.Errorf("getStatusByID query: %w", err)
		}

		batch := make(pushDomain.Pushes, 0, defaultBatchSize)
		for rows.Next() {
			var Push PushDomain.Push
			if err = rows.Scan(&Push.ID, &Push.Status); err != nil {
				rows.Close()
				return nil, fmt.Errorf("getStatusByID scan: %w", err)
			}
			batch = append(batch, Push)
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("getStatusByID rows: %w", err)
		}

		rows.Close()

		if len(batch) == 0 {
			break
		}

		totalpushes = append(totalpushes, batch...)
		lastID = batch[len(batch)-1].ID

		if len(batch) < defaultBatchSize {
			break
		}
	}

	return totalpushes, nil
}

func (r *repository) Deliver(ctx context.Context, messageId string, idempotencyKey uint64) error {
	query := `
        UPDATE emails
        SET status = $1
        AND message_id = $3
        WHERE id = $2
          AND deleted_at IS NULL;`

	result, err := r.pool.Exec(ctx, query, notification.Notification_Status_Delivered, idempotencyKey, messageId)
	if err != nil {
		return fmt.Errorf("deliver query: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("no row updated")
	}

	return nil
}
