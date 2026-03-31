package sms

import (
	"context"
	"insider-one/domain/notification/sms"
)

type GetSmsStatusByIDQuery interface {
	Execute(ctx context.Context, request GetSmsStatusByIDRequest) (*GetSmsStatusByIDResponse, error)
}

type getSmsStatusByIDQuery struct {
	smsRepository sms.Repository
}

func NewGetStatusByIDQuery(smsRepository sms.Repository) GetSmsStatusByIDQuery {
	return &getSmsStatusByIDQuery{smsRepository}
}

func (g *getSmsStatusByIDQuery) Execute(ctx context.Context, request GetSmsStatusByIDRequest) (*GetSmsStatusByIDResponse, error) {
	var ids = []uint64{request.ID}
	getStatusByID, err := g.smsRepository.GetStatusByID(ctx, ids)
	if err != nil {
		return nil, err
	}

	response := GetSmsStatusByIDResponse{
		SmsID:  getStatusByID[0].ID,
		Status: getStatusByID[0].Status,
	}

	return &response, nil
}
