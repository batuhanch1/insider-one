package push

import (
	"context"
	"time"
)

type QueryRepository interface {
	List(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (Pushes, error)
	ListCount(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (int, error)
	GetByStatus(ctx context.Context, status string) ([]uint64, error)
	GetStatusByID(ctx context.Context, ids []uint64) (Pushes, error)
	GetScheduled(ctx context.Context, scheduledAt int64) (Pushes, error)
}

type CommandRepository interface {
	Save(ctx context.Context, push Push) error
	Deliver(ctx context.Context, messageId string, idempotencyKey uint64) error
	Cancel(ctx context.Context, id uint64) error
}
