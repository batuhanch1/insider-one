package email

import (
	"context"
	"errors"
	"testing"

	email_domain "insider-one/domain/notification/email"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelEmailCommand_Execute_Success(t *testing.T) {
	repo := &mockEmailRepository{}
	repo.On("Cancel", mock.Anything, uint64(42)).Return(nil)

	cmd := NewCancelCommand(repo)
	err := cmd.Execute(context.Background(), email_domain.CancelEmailEvent{ID: 42})

	assert.NoError(t, err)
	repo.AssertCalled(t, "Cancel", mock.Anything, uint64(42))
}

func TestCancelEmailCommand_Execute_RepositoryError(t *testing.T) {
	repo := &mockEmailRepository{}
	repo.On("Cancel", mock.Anything, uint64(42)).Return(errors.New("db error"))

	cmd := NewCancelCommand(repo)
	err := cmd.Execute(context.Background(), email_domain.CancelEmailEvent{ID: 42})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
