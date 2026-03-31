package email

import (
	"context"
	"time"
)

type Repository interface {
	Save(ctx context.Context, email Email) error
	List(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (Emails, error)
	ListCount(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (int, error)
	GetByStatus(ctx context.Context, status string) ([]uint64, error)
	UpdateStatus(ctx context.Context, ids []uint64) error
	GetStatusByID(ctx context.Context, ids []uint64) (Emails, error)
	Deliver(ctx context.Context, messageId string, idempotencyKey uint64) error
}
