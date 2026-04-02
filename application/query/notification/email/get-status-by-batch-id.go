package email

import (
	"context"
	"fmt"
	"insider-one/domain/notification/email"
	"insider-one/infrastructure/logging"
)

type GetStatusByBatchIDQuery interface {
	Execute(ctx context.Context, requests GetStatusByBatchIDRequest) (*GetStatusByBatchIDResponse, error)
}

type getStatusByBatchIDQuery struct {
	EmailRepository email.QueryRepository
}

func NewGetStatusByBatchIDQuery(emailRepository email.QueryRepository) GetStatusByBatchIDQuery {
	return &getStatusByBatchIDQuery{emailRepository}
}

func (g *getStatusByBatchIDQuery) Execute(ctx context.Context, requests GetStatusByBatchIDRequest) (*GetStatusByBatchIDResponse, error) {
	getStatusByBatchID, err := g.EmailRepository.GetStatusByID(ctx, requests.IDs)
	if err != nil {
		err = fmt.Errorf("error get email status by ids in getStatusByBatchIDQuery: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	var response GetStatusByBatchIDResponse
	for _, emailByStatus := range getStatusByBatchID {
		response.Emails = append(response.Emails, GetEmailStatusByIDResponse{
			EmailID: emailByStatus.ID,
			Status:  emailByStatus.Status,
		})
	}

	return &response, nil
}
