package push

import (
	"context"
	"errors"
	"testing"

	push_domain "insider-one/domain/notification/push"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPushStatusByIDQuery_Execute_Success(t *testing.T) {
	repo := &mockPushRepository{}
	repo.On("GetStatusByID", mock.Anything, []uint64{5}).Return(
		push_domain.Pushes{{ID: 5, Status: "DELIVERED"}}, nil,
	)

	q := NewGetStatusByIDQuery(repo)
	response, err := q.Execute(context.Background(), GetPushStatusByIDRequest{ID: 5})
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), response.PushID)
	assert.Equal(t, "DELIVERED", response.Status)
}

func TestGetPushStatusByIDQuery_Execute_RepositoryError(t *testing.T) {
	repo := &mockPushRepository{}
	repo.On("GetStatusByID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))

	q := NewGetStatusByIDQuery(repo)
	response, err := q.Execute(context.Background(), GetPushStatusByIDRequest{ID: 1})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetPushStatusByIDQuery_Execute_EmptyResult_Panics(t *testing.T) {
	repo := &mockPushRepository{}
	repo.On("GetStatusByID", mock.Anything, mock.Anything).Return(push_domain.Pushes{}, nil)

	q := NewGetStatusByIDQuery(repo)
	assert.Panics(t, func() {
		_, _ = q.Execute(context.Background(), GetPushStatusByIDRequest{ID: 999})
	})
}
