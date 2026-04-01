package email

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelEmailCommand_Execute_Success_PublishesForEachID(t *testing.T) {
	repo := &mockEmailRepository{}
	pub := &mockPublisher{}

	repo.On("GetByStatus", mock.Anything, "PENDING").Return([]uint64{1, 2, 3}, nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	cmd := NewCancelCommand(repo, pub)
	ctx := context.Background()

	err := cmd.Execute(ctx, CancelEmailRequest{Status: "PENDING"})
	assert.NoError(t, err)
	pub.AssertNumberOfCalls(t, "Publish", 3)
}

func TestCancelEmailCommand_Execute_EmptyIDs_NoPublish(t *testing.T) {
	repo := &mockEmailRepository{}
	pub := &mockPublisher{}

	repo.On("GetByStatus", mock.Anything, "PENDING").Return([]uint64{}, nil)

	cmd := NewCancelCommand(repo, pub)
	ctx := context.Background()

	err := cmd.Execute(ctx, CancelEmailRequest{Status: "PENDING"})
	assert.NoError(t, err)
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCancelEmailCommand_Execute_GetByStatusError(t *testing.T) {
	repo := &mockEmailRepository{}
	pub := &mockPublisher{}

	repo.On("GetByStatus", mock.Anything, "PENDING").Return(nil, errors.New("db error"))

	cmd := NewCancelCommand(repo, pub)
	ctx := context.Background()

	err := cmd.Execute(ctx, CancelEmailRequest{Status: "PENDING"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestCancelEmailCommand_Execute_PublisherError_StopsLoop(t *testing.T) {
	repo := &mockEmailRepository{}
	pub := &mockPublisher{}

	repo.On("GetByStatus", mock.Anything, "PENDING").Return([]uint64{1, 2}, nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("mq error")).Once()

	cmd := NewCancelCommand(repo, pub)
	ctx := context.Background()

	err := cmd.Execute(ctx, CancelEmailRequest{Status: "PENDING"})
	assert.Error(t, err)
	pub.AssertNumberOfCalls(t, "Publish", 1)
}
