package push

import (
	"context"
	"time"
)

type Repository interface {
	Save(ctx context.Context, push Push) error
	List(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (Pushes, error)
	ListCount(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (int, error)
	GetByStatus(ctx context.Context, status string) ([]uint64, error)
	UpdateStatus(ctx context.Context, ids []uint64) error
	GetStatusByID(ctx context.Context, ids []uint64) (Pushes, error)
	Deliver(ctx context.Context, messageId string, idempotencyKey uint64) error
}
