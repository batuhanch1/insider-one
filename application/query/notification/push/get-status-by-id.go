package push

import (
	"context"
	"fmt"
	"insider-one/domain/notification/push"
	"insider-one/infrastructure/logging"
)

type GetPushStatusByIDQuery interface {
	Execute(ctx context.Context, request GetPushStatusByIDRequest) (*GetPushStatusByIDResponse, error)
}

type getPushStatusByIDQuery struct {
	pushRepository push.Repository
}

func NewGetStatusByIDQuery(pushRepository push.Repository) GetPushStatusByIDQuery {
	return &getPushStatusByIDQuery{pushRepository}
}

func (g *getPushStatusByIDQuery) Execute(ctx context.Context, request GetPushStatusByIDRequest) (*GetPushStatusByIDResponse, error) {
	var ids = []uint64{request.ID}
	getStatusByID, err := g.pushRepository.GetStatusByID(ctx, ids)
	if err != nil {
		err = fmt.Errorf("error get push status by id in getPushStatusByIDQuery: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	response := GetPushStatusByIDResponse{
		PushID: getStatusByID[0].ID,
		Status: getStatusByID[0].Status,
	}

	return &response, nil
}
