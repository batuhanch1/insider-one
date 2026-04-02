package email

import (
	"context"
	"fmt"
	"insider-one/domain/notification/email"
	"insider-one/infrastructure/logging"
)

type GetEmailStatusByIDQuery interface {
	Execute(ctx context.Context, request GetEmailStatusByIDRequest) (*GetEmailStatusByIDResponse, error)
}

type getEmailStatusByIDQuery struct {
	EmailRepository email.QueryRepository
}

func NewGetStatusByIDQuery(emailRepository email.QueryRepository) GetEmailStatusByIDQuery {
	return &getEmailStatusByIDQuery{emailRepository}
}

func (g *getEmailStatusByIDQuery) Execute(ctx context.Context, request GetEmailStatusByIDRequest) (*GetEmailStatusByIDResponse, error) {
	var ids = []uint64{request.ID}
	getStatusByID, err := g.EmailRepository.GetStatusByID(ctx, ids)
	if err != nil {
		err = fmt.Errorf("error get email status by id in getEmailStatusByIDQuery: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}
	if len(getStatusByID) == 0 {
		return nil, nil
	}

	response := GetEmailStatusByIDResponse{
		EmailID: getStatusByID[0].ID,
		Status:  getStatusByID[0].Status,
	}

	return &response, nil
}
