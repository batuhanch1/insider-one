package sms

import (
	"context"
	"time"
)

type QueryRepository interface {
	List(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (SmsList, error)
	ListCount(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (int, error)
	GetByStatus(ctx context.Context, status string) ([]uint64, error)
	GetStatusByID(ctx context.Context, ids []uint64) (SmsList, error)
	GetScheduled(ctx context.Context, scheduledAt int64) (SmsList, error)
}

type CommandRepository interface {
	Save(ctx context.Context, sms Sms) error
	Cancel(ctx context.Context, id uint64) error
	Deliver(ctx context.Context, messageId string, idempotencyKey uint64) error
}
