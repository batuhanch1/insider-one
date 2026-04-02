package sms

import (
	"context"
	sms_domain "insider-one/domain/notification/sms"
	sms_provider "insider-one/infrastructure/adapters/client/sms-provider"
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

type mockSmsRepository struct {
	mock.Mock
}

func (m *mockSmsRepository) Save(ctx context.Context, s sms_domain.Sms) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *mockSmsRepository) List(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (sms_domain.SmsList, error) {
	args := m.Called(ctx, status, startDate, endDate, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(sms_domain.SmsList), args.Error(1)
}

func (m *mockSmsRepository) ListCount(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (int, error) {
	args := m.Called(ctx, status, startDate, endDate, page, pageSize)
	return args.Int(0), args.Error(1)
}

func (m *mockSmsRepository) GetByStatus(ctx context.Context, status string) ([]uint64, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uint64), args.Error(1)
}

func (m *mockSmsRepository) Cancel(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSmsRepository) GetStatusByID(ctx context.Context, ids []uint64) (sms_domain.SmsList, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(sms_domain.SmsList), args.Error(1)
}

func (m *mockSmsRepository) Deliver(ctx context.Context, messageId string, idempotencyKey uint64) error {
	args := m.Called(ctx, messageId, idempotencyKey)
	return args.Error(0)
}

func (m *mockSmsRepository) GetScheduled(ctx context.Context, scheduledAt int64) (sms_domain.SmsList, error) {
	args := m.Called(ctx, scheduledAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(sms_domain.SmsList), args.Error(1)
}

type mockSmsProvider struct {
	mock.Mock
}

func (m *mockSmsProvider) Deliver(ctx context.Context, request *sms_provider.DeliverRequest) (*sms_provider.DeliverResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sms_provider.DeliverResponse), args.Error(1)
}
