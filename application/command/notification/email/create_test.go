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

func TestCreateEmailCommand_Execute_NotScheduled_SavesAndPublishes(t *testing.T) {
	repo := &mockEmailRepository{}
	pub := &mockPublisher{}

	repo.On("Save", mock.Anything, mock.Anything).Return(nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
		return opts.Exchange == rabbitmq.Exchange_EmailCreated
	})).Return(nil)

	cmd := NewCreateCommand(repo, pub)
	ctx := context.Background()
	event := email_domain.CreateEmailEvent{
		ScheduledAt: 0, // not scheduled
		To:          "to@a.com",
		From:        "from@a.com",
		Subject:     "Sub",
		Content:     "Body",
		Type:        "T",
		Priority:    "HIGH",
	}

	err := cmd.Execute(ctx, event)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Save", mock.Anything, mock.Anything)
	pub.AssertCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateEmailCommand_Execute_Scheduled_SavesAndSkipsPublish(t *testing.T) {
	repo := &mockEmailRepository{}
	pub := &mockPublisher{}

	repo.On("Save", mock.Anything, mock.Anything).Return(nil)

	cmd := NewCreateCommand(repo, pub)
	ctx := context.Background()
	event := email_domain.CreateEmailEvent{
		ScheduledAt: time.Now().Unix() + 3600, // future = scheduled
		To:          "to@a.com",
		From:        "from@a.com",
		Subject:     "Sub",
		Content:     "Body",
		Type:        "T",
		Priority:    "HIGH",
	}

	err := cmd.Execute(ctx, event)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Save", mock.Anything, mock.Anything)
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateEmailCommand_Execute_RepositoryError_ReturnsError(t *testing.T) {
	repo := &mockEmailRepository{}
	pub := &mockPublisher{}

	repo.On("Save", mock.Anything, mock.Anything).Return(errors.New("db error"))

	cmd := NewCreateCommand(repo, pub)
	ctx := context.Background()
	event := email_domain.CreateEmailEvent{To: "t@a.com", From: "f@a.com"}

	err := cmd.Execute(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateEmailCommand_Execute_PublisherError_ReturnsError(t *testing.T) {
	repo := &mockEmailRepository{}
	pub := &mockPublisher{}

	repo.On("Save", mock.Anything, mock.Anything).Return(nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("mq error"))

	cmd := NewCreateCommand(repo, pub)
	ctx := context.Background()
	event := email_domain.CreateEmailEvent{To: "t@a.com", From: "f@a.com"}

	err := cmd.Execute(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mq error")
}
