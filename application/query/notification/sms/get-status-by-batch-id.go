package sms

import (
	"context"
	"fmt"
	"insider-one/domain/notification/sms"
	"insider-one/infrastructure/logging"
)

type GetStatusByBatchIDQuery interface {
	Execute(ctx context.Context, requests GetStatusByBatchIDRequest) (*GetStatusByBatchIDResponse, error)
}

type getStatusByBatchIDQuery struct {
	smsRepository sms.Repository
}

func NewGetStatusByBatchIDQuery(smsRepository sms.Repository) GetStatusByBatchIDQuery {
	return &getStatusByBatchIDQuery{smsRepository}
}

func (g *getStatusByBatchIDQuery) Execute(ctx context.Context, requests GetStatusByBatchIDRequest) (*GetStatusByBatchIDResponse, error) {
	getStatusByBatchID, err := g.smsRepository.GetStatusByID(ctx, requests.IDs)
	if err != nil {
		err = fmt.Errorf("error get sms status by ids in getStatusByBatchIDQuery: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	var response GetStatusByBatchIDResponse
	for _, smsByStatus := range getStatusByBatchID {
		response.SmsList = append(response.SmsList, GetSmsStatusByIDResponse{
			SmsID:  smsByStatus.ID,
			Status: smsByStatus.Status,
		})
	}

	return &response, nil
}
