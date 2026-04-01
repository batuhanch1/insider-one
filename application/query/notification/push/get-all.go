package push

import (
	"context"
	"fmt"
	"insider-one/domain/notification/push"
	"insider-one/infrastructure/logging"
)

type GetAllQuery interface {
	Execute(ctx context.Context, request GetAllPushRequest) (*GetAllPushResponse, error)
}
type getAllQuery struct {
	pushRepository push.Repository
}

func NewGetAllQuery(pushRepository push.Repository) GetAllQuery {
	return &getAllQuery{pushRepository}
}

func (g *getAllQuery) Execute(ctx context.Context, request GetAllPushRequest) (*GetAllPushResponse, error) {
	pushes, err := g.pushRepository.List(ctx, request.Status, request.CreateDate, request.EndDate, request.Page, request.PageSize)
	if err != nil {
		err = fmt.Errorf("error listing push in get all query: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	filteredPushCount, err := g.pushRepository.ListCount(ctx, request.Status, request.CreateDate, request.EndDate, request.Page, request.PageSize)
	if err != nil {
		err = fmt.Errorf("error listing push count in get all query: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	var response = GetAllPushResponse{
		Pushes:     pushes,
		TotalCount: filteredPushCount,
		Page:       request.Page,
		PageSize:   request.PageSize,
	}

	return &response, nil
}
