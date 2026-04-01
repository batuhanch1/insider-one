package email

import (
	"context"
	"errors"
	"testing"

	email_domain "insider-one/domain/notification/email"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEmailStatusByIDQuery_Execute_Success(t *testing.T) {
	repo := &mockEmailRepository{}
	repo.On("GetStatusByID", mock.Anything, []uint64{42}).Return(
		email_domain.Emails{{ID: 42, Status: "PENDING"}}, nil,
	)

	q := NewGetStatusByIDQuery(repo)
	response, err := q.Execute(context.Background(), GetEmailStatusByIDRequest{ID: 42})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, uint64(42), response.EmailID)
	assert.Equal(t, "PENDING", response.Status)
}

func TestGetEmailStatusByIDQuery_Execute_RepositoryError(t *testing.T) {
	repo := &mockEmailRepository{}
	repo.On("GetStatusByID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))

	q := NewGetStatusByIDQuery(repo)
	response, err := q.Execute(context.Background(), GetEmailStatusByIDRequest{ID: 1})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetEmailStatusByIDQuery_Execute_EmptyResult_Panics(t *testing.T) {
	// Production code accesses getStatusByID[0] without bounds check — panics on empty slice
	repo := &mockEmailRepository{}
	repo.On("GetStatusByID", mock.Anything, mock.Anything).Return(email_domain.Emails{}, nil)

	q := NewGetStatusByIDQuery(repo)
	assert.Panics(t, func() {
		_, _ = q.Execute(context.Background(), GetEmailStatusByIDRequest{ID: 999})
	})
}
