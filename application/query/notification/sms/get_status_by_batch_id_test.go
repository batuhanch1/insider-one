package sms

import (
	"context"
	"errors"
	"testing"

	sms_domain "insider-one/domain/notification/sms"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetStatusByBatchIDSmsQuery_Execute_Success(t *testing.T) {
	repo := &mockSmsRepository{}
	repo.On("GetStatusByID", mock.Anything, []uint64{1, 2}).Return(
		sms_domain.SmsList{
			{ID: 1, Status: "PENDING"},
			{ID: 2, Status: "DELIVERED"},
		}, nil,
	)

	q := NewGetStatusByBatchIDQuery(repo)
	response, err := q.Execute(context.Background(), GetStatusByBatchIDRequest{IDs: []uint64{1, 2}})
	assert.NoError(t, err)
	assert.Len(t, response.SmsList, 2)
	assert.Equal(t, uint64(1), response.SmsList[0].SmsID)
	assert.Equal(t, "PENDING", response.SmsList[0].Status)
}

func TestGetStatusByBatchIDSmsQuery_Execute_RepositoryError(t *testing.T) {
	repo := &mockSmsRepository{}
	repo.On("GetStatusByID", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	q := NewGetStatusByBatchIDQuery(repo)
	response, err := q.Execute(context.Background(), GetStatusByBatchIDRequest{IDs: []uint64{1}})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetStatusByBatchIDSmsQuery_Execute_EmptyResult(t *testing.T) {
	repo := &mockSmsRepository{}
	repo.On("GetStatusByID", mock.Anything, mock.Anything).Return(sms_domain.SmsList{}, nil)

	q := NewGetStatusByBatchIDQuery(repo)
	response, err := q.Execute(context.Background(), GetStatusByBatchIDRequest{IDs: []uint64{}})
	assert.NoError(t, err)
	assert.Nil(t, response.SmsList)
}
