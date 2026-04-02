package email

import (
	"context"
	"time"
)

type CommandRepository interface {
	Save(ctx context.Context, email Email) error
	Cancel(ctx context.Context, id uint64) error
	Deliver(ctx context.Context, messageId string, idempotencyKey uint64) error
}
type QueryRepository interface {
	List(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (Emails, error)
	ListCount(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (int, error)
	GetByStatus(ctx context.Context, status string) ([]uint64, error)
	GetStatusByID(ctx context.Context, ids []uint64) (Emails, error)
	GetScheduled(ctx context.Context, scheduledAt int64) (Emails, error)
}
