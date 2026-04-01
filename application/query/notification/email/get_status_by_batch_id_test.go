package email

import (
	"context"
	"errors"
	"testing"

	email_domain "insider-one/domain/notification/email"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetStatusByBatchIDEmailQuery_Execute_Success(t *testing.T) {
	repo := &mockEmailRepository{}
	repo.On("GetStatusByID", mock.Anything, []uint64{1, 2}).Return(
		email_domain.Emails{
			{ID: 1, Status: "PENDING"},
			{ID: 2, Status: "DELIVERED"},
		}, nil,
	)

	q := NewGetStatusByBatchIDQuery(repo)
	response, err := q.Execute(context.Background(), GetStatusByBatchIDRequest{IDs: []uint64{1, 2}})
	assert.NoError(t, err)
	assert.Len(t, response.Emails, 2)
	assert.Equal(t, uint64(1), response.Emails[0].EmailID)
	assert.Equal(t, "PENDING", response.Emails[0].Status)
	assert.Equal(t, uint64(2), response.Emails[1].EmailID)
	assert.Equal(t, "DELIVERED", response.Emails[1].Status)
}

func TestGetStatusByBatchIDEmailQuery_Execute_RepositoryError(t *testing.T) {
	repo := &mockEmailRepository{}
	repo.On("GetStatusByID", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	q := NewGetStatusByBatchIDQuery(repo)
	response, err := q.Execute(context.Background(), GetStatusByBatchIDRequest{IDs: []uint64{1}})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetStatusByBatchIDEmailQuery_Execute_EmptyResult_ReturnsEmptyResponse(t *testing.T) {
	repo := &mockEmailRepository{}
	repo.On("GetStatusByID", mock.Anything, mock.Anything).Return(email_domain.Emails{}, nil)

	q := NewGetStatusByBatchIDQuery(repo)
	response, err := q.Execute(context.Background(), GetStatusByBatchIDRequest{IDs: []uint64{}})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Nil(t, response.Emails) // for loop never runs, slice stays nil
}
