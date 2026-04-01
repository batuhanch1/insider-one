package push

import (
	"context"
	"errors"
	"testing"

	push_domain "insider-one/domain/notification/push"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetStatusByBatchIDPushQuery_Execute_Success(t *testing.T) {
	repo := &mockPushRepository{}
	repo.On("GetStatusByID", mock.Anything, []uint64{3, 4}).Return(
		push_domain.Pushes{
			{ID: 3, Status: "PENDING"},
			{ID: 4, Status: "DELIVERED"},
		}, nil,
	)

	q := NewGetStatusByBatchIDQuery(repo)
	response, err := q.Execute(context.Background(), GetStatusByBatchIDRequest{IDs: []uint64{3, 4}})
	assert.NoError(t, err)
	assert.Len(t, response.Pushes, 2)
	assert.Equal(t, uint64(3), response.Pushes[0].PushID)
	assert.Equal(t, "PENDING", response.Pushes[0].Status)
}

func TestGetStatusByBatchIDPushQuery_Execute_RepositoryError(t *testing.T) {
	repo := &mockPushRepository{}
	repo.On("GetStatusByID", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	q := NewGetStatusByBatchIDQuery(repo)
	response, err := q.Execute(context.Background(), GetStatusByBatchIDRequest{IDs: []uint64{1}})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetStatusByBatchIDPushQuery_Execute_EmptyResult(t *testing.T) {
	repo := &mockPushRepository{}
	repo.On("GetStatusByID", mock.Anything, mock.Anything).Return(push_domain.Pushes{}, nil)

	q := NewGetStatusByBatchIDQuery(repo)
	response, err := q.Execute(context.Background(), GetStatusByBatchIDRequest{IDs: []uint64{}})
	assert.NoError(t, err)
	assert.Nil(t, response.Pushes)
}
