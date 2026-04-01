package push

import (
	"context"
	"errors"
	"testing"

	push_domain "insider-one/domain/notification/push"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAllPushQuery_Execute_Success(t *testing.T) {
	repo := &mockPushRepository{}
	pushes := push_domain.Pushes{{ID: 1, Status: "PENDING"}}
	repo.On("List", mock.Anything, "PENDING", mock.Anything, mock.Anything, 1, 10).Return(pushes, nil)
	repo.On("ListCount", mock.Anything, "PENDING", mock.Anything, mock.Anything, 1, 10).Return(3, nil)

	q := NewGetAllQuery(repo)
	response, err := q.Execute(context.Background(), GetAllPushRequest{Status: "PENDING", Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, 3, response.TotalCount)
	assert.Len(t, response.Pushes, 1)
}

func TestGetAllPushQuery_Execute_ListError(t *testing.T) {
	repo := &mockPushRepository{}
	repo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("db error"))

	q := NewGetAllQuery(repo)
	response, err := q.Execute(context.Background(), GetAllPushRequest{Status: "PENDING", Page: 1, PageSize: 10})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetAllPushQuery_Execute_ListCountError(t *testing.T) {
	repo := &mockPushRepository{}
	repo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(push_domain.Pushes{}, nil)
	repo.On("ListCount", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(0, errors.New("count error"))

	q := NewGetAllQuery(repo)
	response, err := q.Execute(context.Background(), GetAllPushRequest{Status: "PENDING", Page: 1, PageSize: 10})
	assert.Error(t, err)
	assert.Nil(t, response)
}
