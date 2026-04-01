package push

import (
	"context"
	"errors"
	"testing"
	"time"

	push_domain "insider-one/domain/notification/push"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePushCommand_Execute_NotScheduled_SavesAndPublishes(t *testing.T) {
	repo := &mockPushRepository{}
	pub := &mockPublisher{}

	repo.On("Save", mock.Anything, mock.Anything).Return(nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
		return opts.Exchange == rabbitmq.Exchange_PushCreated
	})).Return(nil)

	cmd := NewCreateCommand(repo, pub)
	ctx := context.Background()
	event := push_domain.CreatePushEvent{
		Sender:      "PUSH",
		PhoneNumber: "+15551234567",
		Type:        "ALERT",
		Content:     "Hello",
		Priority:    "HIGH",
	}

	err := cmd.Execute(ctx, event)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Save", mock.Anything, mock.Anything)
	pub.AssertCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreatePushCommand_Execute_Scheduled_SkipsPublish(t *testing.T) {
	repo := &mockPushRepository{}
	pub := &mockPublisher{}

	repo.On("Save", mock.Anything, mock.Anything).Return(nil)

	cmd := NewCreateCommand(repo, pub)
	ctx := context.Background()
	event := push_domain.CreatePushEvent{
		ScheduledAt: time.Now().Unix() + 3600,
		Sender:      "PUSH",
		PhoneNumber: "+1555",
	}

	err := cmd.Execute(ctx, event)
	assert.NoError(t, err)
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreatePushCommand_Execute_RepositoryError(t *testing.T) {
	repo := &mockPushRepository{}
	pub := &mockPublisher{}

	repo.On("Save", mock.Anything, mock.Anything).Return(errors.New("db error"))

	cmd := NewCreateCommand(repo, pub)
	ctx := context.Background()
	event := push_domain.CreatePushEvent{Sender: "P", PhoneNumber: "+1555"}

	err := cmd.Execute(ctx, event)
	assert.Error(t, err)
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreatePushCommand_Execute_PublisherError(t *testing.T) {
	repo := &mockPushRepository{}
	pub := &mockPublisher{}

	repo.On("Save", mock.Anything, mock.Anything).Return(nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("mq error"))

	cmd := NewCreateCommand(repo, pub)
	ctx := context.Background()
	event := push_domain.CreatePushEvent{Sender: "P", PhoneNumber: "+1555"}

	err := cmd.Execute(ctx, event)
	assert.Error(t, err)
}
