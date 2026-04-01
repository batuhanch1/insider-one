package push

import (
	"context"
	"errors"
	"testing"

	push_domain "insider-one/domain/notification/push"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSchedulePushCommand_Execute_Success_PublishesForEachScheduled(t *testing.T) {
	repo := &mockPushRepository{}
	pub := &mockPublisher{}

	pushes := push_domain.Pushes{
		{ID: 1, Priority: "HIGH"},
		{ID: 2, Priority: "MEDIUM"},
	}
	repo.On("GetScheduled", mock.Anything, mock.Anything).Return(pushes, nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
		return opts.Exchange == rabbitmq.Exchange_PushCreated
	})).Return(nil)

	cmd := NewScheduleCommand(repo, pub)
	err := cmd.Execute(context.Background())
	assert.NoError(t, err)
	pub.AssertNumberOfCalls(t, "Publish", 2)
}

func TestSchedulePushCommand_Execute_NoScheduled(t *testing.T) {
	repo := &mockPushRepository{}
	pub := &mockPublisher{}

	repo.On("GetScheduled", mock.Anything, mock.Anything).Return(push_domain.Pushes{}, nil)

	cmd := NewScheduleCommand(repo, pub)
	err := cmd.Execute(context.Background())
	assert.NoError(t, err)
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestSchedulePushCommand_Execute_RepoError(t *testing.T) {
	repo := &mockPushRepository{}
	pub := &mockPublisher{}

	repo.On("GetScheduled", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	cmd := NewScheduleCommand(repo, pub)
	err := cmd.Execute(context.Background())
	assert.Error(t, err)
}

func TestSchedulePushCommand_Execute_PublisherError_ReturnsNil(t *testing.T) {
	repo := &mockPushRepository{}
	pub := &mockPublisher{}

	repo.On("GetScheduled", mock.Anything, mock.Anything).Return(push_domain.Pushes{{ID: 1, Priority: "LOW"}}, nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("mq error"))

	cmd := NewScheduleCommand(repo, pub)
	err := cmd.Execute(context.Background())
	assert.NoError(t, err) // schedule command swallows publisher errors
}
