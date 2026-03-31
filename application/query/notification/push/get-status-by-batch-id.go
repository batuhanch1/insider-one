package push

import (
	"context"
	"insider-one/domain/notification/push"
)

type GetStatusByBatchIDQuery interface {
	Execute(ctx context.Context, requests GetStatusByBatchIDRequest) (*GetStatusByBatchIDResponse, error)
}

type getStatusByBatchIDQuery struct {
	pushRepository push.Repository
}

func NewGetStatusByBatchIDQuery(pushRepository push.Repository) GetStatusByBatchIDQuery {
	return &getStatusByBatchIDQuery{pushRepository}
}

func (g *getStatusByBatchIDQuery) Execute(ctx context.Context, requests GetStatusByBatchIDRequest) (*GetStatusByBatchIDResponse, error) {
	getStatusByBatchID, err := g.pushRepository.GetStatusByID(ctx, requests.IDs)
	if err != nil {
		return nil, err
	}

	var response GetStatusByBatchIDResponse
	for _, pushByStatus := range getStatusByBatchID {
		response.Pushes = append(response.Pushes, GetPushStatusByIDResponse{
			PushID: pushByStatus.ID,
			Status: pushByStatus.Status,
		})
	}

	return &response, nil
}
