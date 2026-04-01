package push

import (
	"context"
	push_domain "insider-one/domain/notification/push"
	push_provider "insider-one/infrastructure/adapters/client/push-provider"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"time"

	"github.com/stretchr/testify/mock"
)

type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) Publish(ctx context.Context, v any, opts rabbitmq.PublishOptions) error {
	args := m.Called(ctx, v, opts)
	return args.Error(0)
}

type mockBatchPublisher struct {
	mock.Mock
}

func (m *mockBatchPublisher) Publish(ctx context.Context, messages []any, opts rabbitmq.BatchPublisherOptions) error {
	args := m.Called(ctx, messages, opts)
	return args.Error(0)
}

type mockPushRepository struct {
	mock.Mock
}

func (m *mockPushRepository) Save(ctx context.Context, p push_domain.Push) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *mockPushRepository) List(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (push_domain.Pushes, error) {
	args := m.Called(ctx, status, startDate, endDate, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(push_domain.Pushes), args.Error(1)
}

func (m *mockPushRepository) ListCount(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (int, error) {
	args := m.Called(ctx, status, startDate, endDate, page, pageSize)
	return args.Int(0), args.Error(1)
}

func (m *mockPushRepository) GetByStatus(ctx context.Context, status string) ([]uint64, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uint64), args.Error(1)
}

func (m *mockPushRepository) UpdateStatus(ctx context.Context, ids []uint64) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *mockPushRepository) GetStatusByID(ctx context.Context, ids []uint64) (push_domain.Pushes, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(push_domain.Pushes), args.Error(1)
}

func (m *mockPushRepository) Deliver(ctx context.Context, messageId string, idempotencyKey uint64) error {
	args := m.Called(ctx, messageId, idempotencyKey)
	return args.Error(0)
}

func (m *mockPushRepository) GetScheduled(ctx context.Context, scheduledAt int64) (push_domain.Pushes, error) {
	args := m.Called(ctx, scheduledAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(push_domain.Pushes), args.Error(1)
}

type mockPushProvider struct {
	mock.Mock
}

func (m *mockPushProvider) Deliver(ctx context.Context, request *push_provider.DeliverRequest) (*push_provider.DeliverResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*push_provider.DeliverResponse), args.Error(1)
}
