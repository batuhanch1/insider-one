package sms

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelSmsCommand_Execute_Success_PublishesForEachID(t *testing.T) {
	repo := &mockSmsRepository{}
	pub := &mockPublisher{}

	repo.On("GetByStatus", mock.Anything, "PENDING").Return([]uint64{1, 2}, nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	cmd := NewEnqueueCancelCommand(repo, pub)
	err := cmd.Execute(context.Background(), CancelSmsRequest{Status: "PENDING"})
	assert.NoError(t, err)
	pub.AssertNumberOfCalls(t, "Publish", 2)
}

func TestCancelSmsCommand_Execute_EmptyIDs_NoPublish(t *testing.T) {
	repo := &mockSmsRepository{}
	pub := &mockPublisher{}

	repo.On("GetByStatus", mock.Anything, "PENDING").Return([]uint64{}, nil)

	cmd := NewEnqueueCancelCommand(repo, pub)
	err := cmd.Execute(context.Background(), CancelSmsRequest{Status: "PENDING"})
	assert.NoError(t, err)
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCancelSmsCommand_Execute_GetByStatusError(t *testing.T) {
	repo := &mockSmsRepository{}
	pub := &mockPublisher{}

	repo.On("GetByStatus", mock.Anything, "PENDING").Return(nil, errors.New("db error"))

	cmd := NewEnqueueCancelCommand(repo, pub)
	err := cmd.Execute(context.Background(), CancelSmsRequest{Status: "PENDING"})
	assert.Error(t, err)
}

func TestCancelSmsCommand_Execute_PublisherError_StopsLoop(t *testing.T) {
	repo := &mockSmsRepository{}
	pub := &mockPublisher{}

	repo.On("GetByStatus", mock.Anything, "PENDING").Return([]uint64{1, 2}, nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("mq error")).Once()

	cmd := NewEnqueueCancelCommand(repo, pub)
	err := cmd.Execute(context.Background(), CancelSmsRequest{Status: "PENDING"})
	assert.Error(t, err)
	pub.AssertNumberOfCalls(t, "Publish", 1)
}
