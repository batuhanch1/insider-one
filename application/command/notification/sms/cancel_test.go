package sms

import (
	"context"
	"errors"
	"testing"

	sms_domain "insider-one/domain/notification/sms"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelSmsCommand_Execute_Success(t *testing.T) {
	repo := &mockSmsRepository{}
	repo.On("Cancel", mock.Anything, uint64(42)).Return(nil)

	cmd := NewCancelCommand(repo)
	err := cmd.Execute(context.Background(), sms_domain.CancelSmsEvent{ID: 42})

	assert.NoError(t, err)
	repo.AssertCalled(t, "Cancel", mock.Anything, uint64(42))
}

func TestCancelSmsCommand_Execute_RepositoryError(t *testing.T) {
	repo := &mockSmsRepository{}
	repo.On("Cancel", mock.Anything, uint64(42)).Return(errors.New("db error"))

	cmd := NewCancelCommand(repo)
	err := cmd.Execute(context.Background(), sms_domain.CancelSmsEvent{ID: 42})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
