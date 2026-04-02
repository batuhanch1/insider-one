package email

import (
	"context"
	email_domain "insider-one/domain/notification/email"
	email_provider "insider-one/infrastructure/adapters/client/email-provider"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"time"

	"github.com/stretchr/testify/mock"
)

// mockPublisher implements the publisher interface
type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) Publish(ctx context.Context, v any, opts rabbitmq.PublishOptions) error {
	args := m.Called(ctx, v, opts)
	return args.Error(0)
}

// mockBatchPublisher implements the batchPublisher interface
type mockBatchPublisher struct {
	mock.Mock
}

func (m *mockBatchPublisher) Publish(ctx context.Context, messages []any, opts rabbitmq.BatchPublisherOptions) error {
	args := m.Called(ctx, messages, opts)
	return args.Error(0)
}

// mockEmailRepository implements email_domain.Repository
type mockEmailRepository struct {
	mock.Mock
}

func (m *mockEmailRepository) Save(ctx context.Context, e email_domain.Email) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *mockEmailRepository) List(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (email_domain.Emails, error) {
	args := m.Called(ctx, status, startDate, endDate, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(email_domain.Emails), args.Error(1)
}

func (m *mockEmailRepository) ListCount(ctx context.Context, status string, startDate, endDate *time.Time, page, pageSize int) (int, error) {
	args := m.Called(ctx, status, startDate, endDate, page, pageSize)
	return args.Int(0), args.Error(1)
}

func (m *mockEmailRepository) GetByStatus(ctx context.Context, status string) ([]uint64, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uint64), args.Error(1)
}

func (m *mockEmailRepository) Cancel(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockEmailRepository) GetStatusByID(ctx context.Context, ids []uint64) (email_domain.Emails, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(email_domain.Emails), args.Error(1)
}

func (m *mockEmailRepository) Deliver(ctx context.Context, messageId string, idempotencyKey uint64) error {
	args := m.Called(ctx, messageId, idempotencyKey)
	return args.Error(0)
}

func (m *mockEmailRepository) GetScheduled(ctx context.Context, scheduledAt int64) (email_domain.Emails, error) {
	args := m.Called(ctx, scheduledAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(email_domain.Emails), args.Error(1)
}

// mockEmailProvider implements email_provider.EmailProvider
type mockEmailProvider struct {
	mock.Mock
}

func (m *mockEmailProvider) Deliver(ctx context.Context, request *email_provider.DeliverRequest) (*email_provider.DeliverResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*email_provider.DeliverResponse), args.Error(1)
}
