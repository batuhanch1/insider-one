package push

import (
	"context"
	"errors"
	"testing"

	push_domain "insider-one/domain/notification/push"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelPushCommand_Execute_Success(t *testing.T) {
	repo := &mockPushRepository{}
	repo.On("Cancel", mock.Anything, uint64(42)).Return(nil)

	cmd := NewCancelCommand(repo)
	err := cmd.Execute(context.Background(), push_domain.CancelPushEvent{ID: 42})

	assert.NoError(t, err)
	repo.AssertCalled(t, "Cancel", mock.Anything, uint64(42))
}

func TestCancelPushCommand_Execute_RepositoryError(t *testing.T) {
	repo := &mockPushRepository{}
	repo.On("Cancel", mock.Anything, uint64(42)).Return(errors.New("db error"))

	cmd := NewCancelCommand(repo)
	err := cmd.Execute(context.Background(), push_domain.CancelPushEvent{ID: 42})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
