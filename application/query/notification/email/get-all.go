package email

import (
	"context"
	"fmt"
	"insider-one/domain/notification/email"
	"insider-one/infrastructure/logging"
)

type GetAllQuery interface {
	Execute(ctx context.Context, request GetAllEmailRequest) (*GetAllEmailResponse, error)
}
type getAllQuery struct {
	EmailRepository email.QueryRepository
}

func NewGetAllQuery(emailRepository email.QueryRepository) GetAllQuery {
	return &getAllQuery{emailRepository}
}

func (g *getAllQuery) Execute(ctx context.Context, request GetAllEmailRequest) (*GetAllEmailResponse, error) {
	emails, err := g.EmailRepository.List(ctx, request.Status, request.CreateDate, request.EndDate, request.Page, request.PageSize)
	if err != nil {
		err = fmt.Errorf("error listing email in get all query: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	filteredEmailCount, err := g.EmailRepository.ListCount(ctx, request.Status, request.CreateDate, request.EndDate, request.Page, request.PageSize)
	if err != nil {
		err = fmt.Errorf("error listing email count in get all query: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	var response = GetAllEmailResponse{
		Emails:     emails,
		TotalCount: filteredEmailCount,
		Page:       request.Page,
		PageSize:   request.PageSize,
	}

	return &response, nil
}
