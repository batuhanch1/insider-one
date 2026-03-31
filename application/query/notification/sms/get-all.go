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
	SmsRepository sms.Repository
}

func NewGetAllQuery(SmsRepository sms.Repository) GetAllQuery {
	return &getAllQuery{SmsRepository}
}

func (g *getAllQuery) Execute(ctx context.Context, request GetAllSmsRequest) (*GetAllSmsResponse, error) {
	smsList, err := g.SmsRepository.List(ctx, request.Status, request.CreateDate, request.EndDate, request.Page, request.PageSize)
	if err != nil {
		err = fmt.Errorf("Execute.SmsRepository.List: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	filteredSmsCount, err := g.SmsRepository.ListCount(ctx, request.Status, request.CreateDate, request.EndDate, request.Page, request.PageSize)
	if err != nil {
		err = fmt.Errorf("Execute.SmsRepository.ListCount: %w", err)
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
