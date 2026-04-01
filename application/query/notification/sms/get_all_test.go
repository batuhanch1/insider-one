package sms

import (
	"context"
	"errors"
	"testing"

	sms_domain "insider-one/domain/notification/sms"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAllSmsQuery_Execute_Success(t *testing.T) {
	repo := &mockSmsRepository{}
	smsList := sms_domain.SmsList{{ID: 1, Status: "PENDING"}, {ID: 2, Status: "PENDING"}}
	repo.On("List", mock.Anything, "PENDING", mock.Anything, mock.Anything, 1, 10).Return(smsList, nil)
	repo.On("ListCount", mock.Anything, "PENDING", mock.Anything, mock.Anything, 1, 10).Return(5, nil)

	q := NewGetAllQuery(repo)
	response, err := q.Execute(context.Background(), GetAllSmsRequest{Status: "PENDING", Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, 5, response.TotalCount)
	assert.Len(t, response.SmsList, 2)
}

func TestGetAllSmsQuery_Execute_ListError(t *testing.T) {
	repo := &mockSmsRepository{}
	repo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("db error"))

	q := NewGetAllQuery(repo)
	response, err := q.Execute(context.Background(), GetAllSmsRequest{Status: "PENDING", Page: 1, PageSize: 10})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetAllSmsQuery_Execute_ListCountError(t *testing.T) {
	repo := &mockSmsRepository{}
	repo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(sms_domain.SmsList{}, nil)
	repo.On("ListCount", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(0, errors.New("count error"))

	q := NewGetAllQuery(repo)
	response, err := q.Execute(context.Background(), GetAllSmsRequest{Status: "PENDING", Page: 1, PageSize: 10})
	assert.Error(t, err)
	assert.Nil(t, response)
}
