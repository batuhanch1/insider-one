package email

import (
	"context"
	"errors"
	"testing"

	email_domain "insider-one/domain/notification/email"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAllEmailQuery_Execute_Success(t *testing.T) {
	repo := &mockEmailRepository{}
	emails := email_domain.Emails{
		{ID: 1, Status: "PENDING"},
		{ID: 2, Status: "PENDING"},
	}
	repo.On("List", mock.Anything, "PENDING", mock.Anything, mock.Anything, 1, 10).Return(emails, nil)
	repo.On("ListCount", mock.Anything, "PENDING", mock.Anything, mock.Anything, 1, 10).Return(20, nil)

	q := NewGetAllQuery(repo)
	ctx := context.Background()
	req := GetAllEmailRequest{Status: "PENDING", Page: 1, PageSize: 10}

	response, err := q.Execute(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 20, response.TotalCount)
	assert.Len(t, response.Emails, 2)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PageSize)
}

func TestGetAllEmailQuery_Execute_ListError(t *testing.T) {
	repo := &mockEmailRepository{}
	repo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("db error"))

	q := NewGetAllQuery(repo)
	response, err := q.Execute(context.Background(), GetAllEmailRequest{Status: "PENDING", Page: 1, PageSize: 10})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetAllEmailQuery_Execute_ListCountError(t *testing.T) {
	repo := &mockEmailRepository{}
	repo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(email_domain.Emails{}, nil)
	repo.On("ListCount", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(0, errors.New("count error"))

	q := NewGetAllQuery(repo)
	response, err := q.Execute(context.Background(), GetAllEmailRequest{Status: "PENDING", Page: 1, PageSize: 10})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetAllEmailQuery_Execute_EmptyResults(t *testing.T) {
	repo := &mockEmailRepository{}
	repo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(email_domain.Emails{}, nil)
	repo.On("ListCount", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(0, nil)

	q := NewGetAllQuery(repo)
	response, err := q.Execute(context.Background(), GetAllEmailRequest{Status: "PENDING", Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, 0, response.TotalCount)
}
