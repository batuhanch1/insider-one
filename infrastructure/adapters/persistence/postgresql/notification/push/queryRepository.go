package push

import (
	"context"
	"database/sql"
	"fmt"
	"insider-one/domain/notification"
	PushDomain "insider-one/domain/notification/push"
	pushDomain "insider-one/domain/notification/push"
	"insider-one/infrastructure/logging"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultBatchSize = 100

type queryRepository struct {
	pool *pgxpool.Pool
}

func NewQueryRepository(pool *pgxpool.Pool) pushDomain.QueryRepository {
	return &queryRepository{pool}
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

func (r *queryRepository) List(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (pushDomain.Pushes, error) {
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

	logging.DbQueryStart(ctx, query)
	rows, err := r.pool.Query(ctx, query, args...)
	logging.DbQueryFinish(ctx)
	if err != nil {
		err = fmt.Errorf("list push query error: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}
	defer rows.Close()

	var pushes pushDomain.Pushes
	for rows.Next() {
		var push PushDomain.Push
		var scheduledAt, sentAt, deletedAt sql.NullInt64

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
			err = fmt.Errorf("list push scan error: %w", err)
			logging.Error(ctx, err)
			return nil, err
		}

		if scheduledAt.Valid {
			push.ScheduledAt = &scheduledAt.Int64
		}
		if sentAt.Valid {
			push.SentAt = &sentAt.Int64
		}
		if deletedAt.Valid {
			push.DeletedAt = &deletedAt.Int64
		}

		pushes = append(pushes, push)
	}

	if err = rows.Err(); err != nil {
		err = fmt.Errorf("list push rows error: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}
	return pushes, nil
}

func (r *queryRepository) ListCount(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (int, error) {
	where, args, _, _ := buildFilter(status, startDate, endDate, page, pageSize)

	query := fmt.Sprintf(`
        SELECT
            count(1) as totalCount
        FROM pushes
        %s
    `, where)

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

func (r *queryRepository) GetByStatus(ctx context.Context, status string) ([]uint64, error) {
	var lastID uint64 = 0
	var totalIDs []uint64

	query := `
			SELECT id FROM pushes
			WHERE status = $1
			  AND deleted_at IS NULL
			  AND id > $2
			ORDER BY id ASC
			LIMIT $3
		`
	for {
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

func (r *queryRepository) GetStatusByID(ctx context.Context, ids []uint64) (pushDomain.Pushes, error) {
	var lastID uint64 = 0
	var totalpushes pushDomain.Pushes

	query := `
			SELECT id, status FROM pushes
			WHERE id = ANY($1)
			  AND deleted_at IS NULL
			  AND id > $2
			ORDER BY id ASC
			LIMIT $3
		`
	for {
		logging.DbQueryStart(ctx, query)
		rows, err := r.pool.Query(ctx, query, ids, lastID, defaultBatchSize)
		logging.DbQueryFinish(ctx)
		if err != nil {
			err = fmt.Errorf("getStatusByID query: %w", err)
			logging.Error(ctx, err)
			return nil, err
		}

		batch := make(pushDomain.Pushes, 0, defaultBatchSize)
		for rows.Next() {
			var Push PushDomain.Push
			if err = rows.Scan(&Push.ID, &Push.Status); err != nil {
				rows.Close()
				err = fmt.Errorf("getStatusByID scan: %w", err)
				logging.Error(ctx, err)
				return nil, err
			}
			batch = append(batch, Push)
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

		totalpushes = append(totalpushes, batch...)
		lastID = batch[len(batch)-1].ID

		if len(batch) < defaultBatchSize {
			break
		}
	}

	return totalpushes, nil
}

func (r *queryRepository) GetScheduled(ctx context.Context, scheduledAt int64) (pushDomain.Pushes, error) {
	var lastID uint64 = 0
	var totalPushes pushDomain.Pushes

	query := `
			SELECT id,scheduled_at,idempotency_key,sender,phone_number,type,status,content,priority FROM pushes
			WHERE status = $1
			  AND deleted_at IS NULL
			  AND scheduled_at < $4
			  AND id > $2
			ORDER BY id ASC
			LIMIT $3
		`
	for {
		logging.DbQueryStart(ctx, query)
		rows, err := r.pool.Query(ctx, query, notification.Notification_Status_Scheduled, lastID, defaultBatchSize, scheduledAt)
		logging.DbQueryFinish(ctx)
		if err != nil {
			err = fmt.Errorf("getScheduled query: %w", err)
			logging.Error(ctx, err)
			return nil, err
		}

		pushes := make(pushDomain.Pushes, 0, defaultBatchSize)
		for rows.Next() {
			var e pushDomain.Push
			if err = rows.Scan(&e.ID, &e.ScheduledAt, &e.IdempotencyKey, &e.Sender, &e.PhoneNumber, &e.Type, &e.Status, &e.Content, &e.Priority); err != nil {
				rows.Close()
				err = fmt.Errorf("getScheduled scan: %w", err)
				logging.Error(ctx, err)
				return nil, err
			}
			pushes = append(pushes, e)
		}

		if err = rows.Err(); err != nil {
			rows.Close()
			err = fmt.Errorf("getScheduled rows: %w", err)
			logging.Error(ctx, err)
			return nil, err
		}

		rows.Close()

		if len(pushes) == 0 {
			break
		}

		totalPushes = append(totalPushes, pushes...)
		lastID = pushes[len(pushes)-1].ID

		if len(pushes) < defaultBatchSize {
			break
		}
	}

	return totalPushes, nil
}
