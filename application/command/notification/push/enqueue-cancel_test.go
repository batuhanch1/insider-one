package push

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelPushCommand_Execute_Success_PublishesForEachID(t *testing.T) {
	repo := &mockPushRepository{}
	pub := &mockPublisher{}

	repo.On("GetByStatus", mock.Anything, "PENDING").Return([]uint64{10, 20}, nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	cmd := NewEnqueueCancelCommand(repo, pub)
	err := cmd.Execute(context.Background(), CancelPushRequest{Status: "PENDING"})
	assert.NoError(t, err)
	pub.AssertNumberOfCalls(t, "Publish", 2)
}

func TestCancelPushCommand_Execute_EmptyIDs(t *testing.T) {
	repo := &mockPushRepository{}
	pub := &mockPublisher{}

	repo.On("GetByStatus", mock.Anything, "PENDING").Return([]uint64{}, nil)

	cmd := NewEnqueueCancelCommand(repo, pub)
	err := cmd.Execute(context.Background(), CancelPushRequest{Status: "PENDING"})
	assert.NoError(t, err)
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCancelPushCommand_Execute_GetByStatusError(t *testing.T) {
	repo := &mockPushRepository{}
	pub := &mockPublisher{}

	repo.On("GetByStatus", mock.Anything, "PENDING").Return(nil, errors.New("db error"))

	cmd := NewEnqueueCancelCommand(repo, pub)
	err := cmd.Execute(context.Background(), CancelPushRequest{Status: "PENDING"})
	assert.Error(t, err)
}

func TestCancelPushCommand_Execute_PublisherError_StopsLoop(t *testing.T) {
	repo := &mockPushRepository{}
	pub := &mockPublisher{}

	repo.On("GetByStatus", mock.Anything, "PENDING").Return([]uint64{1, 2}, nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("mq error")).Once()

	cmd := NewEnqueueCancelCommand(repo, pub)
	err := cmd.Execute(context.Background(), CancelPushRequest{Status: "PENDING"})
	assert.Error(t, err)
	pub.AssertNumberOfCalls(t, "Publish", 1)
}
