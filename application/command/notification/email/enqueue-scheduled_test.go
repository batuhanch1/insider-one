package email

import (
	"context"
	"errors"
	"testing"
	"time"

	email_domain "insider-one/domain/notification/email"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestScheduleEmailCommand_Execute_Success_PublishesForEachScheduled(t *testing.T) {
	repo := &mockEmailRepository{}
	pub := &mockPublisher{}

	emails := email_domain.Emails{
		{ID: 1, Priority: "HIGH", To: "a@a.com", From: "b@b.com"},
		{ID: 2, Priority: "LOW", To: "c@c.com", From: "d@d.com"},
	}
	repo.On("GetScheduled", mock.Anything, mock.Anything).Return(emails, nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
		return opts.Exchange == rabbitmq.Exchange_EmailCreated
	})).Return(nil)

	cmd := NewScheduleCommand(repo, pub)
	ctx := context.Background()

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	pub.AssertNumberOfCalls(t, "Publish", 2)
}

func TestScheduleEmailCommand_Execute_NoScheduledEmails(t *testing.T) {
	repo := &mockEmailRepository{}
	pub := &mockPublisher{}

	repo.On("GetScheduled", mock.Anything, mock.Anything).Return(email_domain.Emails{}, nil)

	cmd := NewScheduleCommand(repo, pub)
	ctx := context.Background()

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestScheduleEmailCommand_Execute_RepoError(t *testing.T) {
	repo := &mockEmailRepository{}
	pub := &mockPublisher{}

	repo.On("GetScheduled", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	cmd := NewScheduleCommand(repo, pub)
	ctx := context.Background()

	err := cmd.Execute(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestScheduleEmailCommand_Execute_PublisherError_ReturnsNil(t *testing.T) {
	// The schedule command swallows publisher errors and always returns nil
	repo := &mockEmailRepository{}
	pub := &mockPublisher{}

	emails := email_domain.Emails{
		{ID: 1, Priority: "HIGH", ScheduledAt: time.Now().Unix() + 300},
	}
	repo.On("GetScheduled", mock.Anything, mock.Anything).Return(emails, nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("mq error"))

	cmd := NewScheduleCommand(repo, pub)
	ctx := context.Background()

	err := cmd.Execute(ctx)
	assert.NoError(t, err) // publisher errors are logged and swallowed
}
