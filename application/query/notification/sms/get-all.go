package sms

import (
	"context"
	"fmt"
	"insider-one/domain/notification/sms"
	"insider-one/infrastructure/logging"
)

type GetAllQuery interface {
	Execute(ctx context.Context, request GetAllSmsRequest) (*GetAllSmsResponse, error)
}
type getAllQuery struct {
	SmsRepository sms.QueryRepository
}

func NewGetAllQuery(SmsRepository sms.QueryRepository) GetAllQuery {
	return &getAllQuery{SmsRepository}
}

func (g *getAllQuery) Execute(ctx context.Context, request GetAllSmsRequest) (*GetAllSmsResponse, error) {
	smsList, err := g.SmsRepository.List(ctx, request.Status, request.CreateDate, request.EndDate, request.Page, request.PageSize)
	if err != nil {
		err = fmt.Errorf("error listing sms in get all query: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	filteredSmsCount, err := g.SmsRepository.ListCount(ctx, request.Status, request.CreateDate, request.EndDate, request.Page, request.PageSize)
	if err != nil {
		err = fmt.Errorf("error listing sms count in get all query: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	var response = GetAllSmsResponse{
		SmsList:    smsList,
		TotalCount: filteredSmsCount,
		Page:       request.Page,
		PageSize:   request.PageSize,
	}

	return &response, nil
}
