package sms

import (
	"context"
	"errors"
	"testing"

	sms_domain "insider-one/domain/notification/sms"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSmsStatusByIDQuery_Execute_Success(t *testing.T) {
	repo := &mockSmsRepository{}
	repo.On("GetStatusByID", mock.Anything, []uint64{10}).Return(
		sms_domain.SmsList{{ID: 10, Status: "PENDING"}}, nil,
	)

	q := NewGetStatusByIDQuery(repo)
	response, err := q.Execute(context.Background(), GetSmsStatusByIDRequest{ID: 10})
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), response.SmsID)
	assert.Equal(t, "PENDING", response.Status)
}

func TestGetSmsStatusByIDQuery_Execute_RepositoryError(t *testing.T) {
	repo := &mockSmsRepository{}
	repo.On("GetStatusByID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))

	q := NewGetStatusByIDQuery(repo)
	response, err := q.Execute(context.Background(), GetSmsStatusByIDRequest{ID: 1})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetSmsStatusByIDQuery_Execute_EmptyResult_Panics(t *testing.T) {
	repo := &mockSmsRepository{}
	repo.On("GetStatusByID", mock.Anything, mock.Anything).Return(sms_domain.SmsList{}, nil)

	q := NewGetStatusByIDQuery(repo)
	assert.Panics(t, func() {
		_, _ = q.Execute(context.Background(), GetSmsStatusByIDRequest{ID: 999})
	})
}
