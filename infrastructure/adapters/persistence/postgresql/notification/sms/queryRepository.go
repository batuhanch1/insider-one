package sms

import (
	"context"
	"database/sql"
	"fmt"
	"insider-one/domain/notification"
	smsDomain "insider-one/domain/notification/sms"
	"insider-one/infrastructure/logging"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultBatchSize = 100

type queryRepository struct {
	pool *pgxpool.Pool
}

func NewQueryRepository(pool *pgxpool.Pool) smsDomain.QueryRepository {
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

func (r *queryRepository) List(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (smsDomain.SmsList, error) {
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
		var scheduledAt, sentAt, deletedAt sql.NullInt64

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
			sms.ScheduledAt = &scheduledAt.Int64
		}
		if sentAt.Valid {
			sms.SentAt = &sentAt.Int64
		}
		if deletedAt.Valid {
			sms.DeletedAt = &deletedAt.Int64
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

func (r *queryRepository) ListCount(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (int, error) {
	where, args, _, _ := buildFilter(status, startDate, endDate, page, pageSize)

	query := fmt.Sprintf(`
        SELECT
            count(1) as totalCount
        FROM sms
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
			SELECT id FROM sms
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

func (r *queryRepository) GetStatusByID(ctx context.Context, ids []uint64) (smsDomain.SmsList, error) {
	var lastID uint64 = 0
	var totalSmsList smsDomain.SmsList

	query := `
			SELECT id, status FROM sms
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

func (r *queryRepository) GetScheduled(ctx context.Context, scheduledAt int64) (smsDomain.SmsList, error) {
	var lastID uint64 = 0
	var totalSmsList smsDomain.SmsList

	query := `
			SELECT id,scheduled_at,idempotency_key,phone_number,sender,type,status,content,priority FROM sms
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

		smsList := make(smsDomain.SmsList, 0, defaultBatchSize)
		for rows.Next() {
			var e smsDomain.Sms
			if err = rows.Scan(&e.ID, &e.ScheduledAt, &e.IdempotencyKey, &e.PhoneNumber, &e.Sender, &e.Type, &e.Status, &e.Content, &e.Priority); err != nil {
				rows.Close()
				err = fmt.Errorf("getScheduled scan: %w", err)
				logging.Error(ctx, err)
				return nil, err
			}
			smsList = append(smsList, e)
		}

		if err = rows.Err(); err != nil {
			rows.Close()
			err = fmt.Errorf("getScheduled rows: %w", err)
			logging.Error(ctx, err)
			return nil, err
		}

		rows.Close()

		if len(smsList) == 0 {
			break
		}

		totalSmsList = append(totalSmsList, smsList...)
		lastID = smsList[len(smsList)-1].ID

		if len(smsList) < defaultBatchSize {
			break
		}
	}

	return totalSmsList, nil
}
