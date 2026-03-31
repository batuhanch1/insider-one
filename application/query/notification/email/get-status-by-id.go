package email

import (
	"context"
	"insider-one/domain/notification/email"
)

type GetEmailStatusByIDQuery interface {
	Execute(ctx context.Context, request GetEmailStatusByIDRequest) (*GetEmailStatusByIDResponse, error)
}

type getEmailStatusByIDQuery struct {
	EmailRepository email.Repository
}

func NewGetStatusByIDQuery(emailRepository email.Repository) GetEmailStatusByIDQuery {
	return &getEmailStatusByIDQuery{emailRepository}
}

func (g *getEmailStatusByIDQuery) Execute(ctx context.Context, request GetEmailStatusByIDRequest) (*GetEmailStatusByIDResponse, error) {
	var ids = []uint64{request.ID}
	getStatusByID, err := g.EmailRepository.GetStatusByID(ctx, ids)
	if err != nil {
		return nil, err
	}

	response := GetEmailStatusByIDResponse{
		EmailID: getStatusByID[0].ID,
		Status:  getStatusByID[0].Status,
	}

	return &response, nil
}
