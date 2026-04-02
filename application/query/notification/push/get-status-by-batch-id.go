package push

import (
	"context"
	"fmt"
	"insider-one/domain/notification/push"
	"insider-one/infrastructure/logging"
)

type GetStatusByBatchIDQuery interface {
	Execute(ctx context.Context, requests GetStatusByBatchIDRequest) (*GetStatusByBatchIDResponse, error)
}

type getStatusByBatchIDQuery struct {
	pushRepository push.QueryRepository
}

func NewGetStatusByBatchIDQuery(pushRepository push.QueryRepository) GetStatusByBatchIDQuery {
	return &getStatusByBatchIDQuery{pushRepository}
}

func (g *getStatusByBatchIDQuery) Execute(ctx context.Context, requests GetStatusByBatchIDRequest) (*GetStatusByBatchIDResponse, error) {
	getStatusByBatchID, err := g.pushRepository.GetStatusByID(ctx, requests.IDs)
	if err != nil {
		err = fmt.Errorf("error get push status by ids in getStatusByBatchIDQuery: %w", err)
		logging.Error(ctx, err)
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
