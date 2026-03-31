package email

import (
	"context"
	"fmt"
	"insider-one/domain/notification/email"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"
)

type CancelCommand interface {
	Execute(ctx context.Context, request CancelEmailRequest) error
}

type cancelCommand struct {
	EmailRepository email.Repository
	Publisher       rabbitmq2.Publisher
}

func NewCancelCommand(emailRepository email.Repository, publisher rabbitmq2.Publisher) CancelCommand {
	return &cancelCommand{emailRepository, publisher}
}

func (s *cancelCommand) Execute(ctx context.Context, request CancelEmailRequest) error {
	emailIds, err := s.EmailRepository.GetByStatus(ctx, request.Status)
	if err != nil {
		return fmt.Errorf("cancelPendingCommand.emailRepository.GetByStatus: %w", err)
	}

	for _, emailId := range emailIds {
		cancelEmailEvent := email.CancelEmailEvent{ID: emailId}
		err = s.Publisher.Publish(ctx, cancelEmailEvent, rabbitmq2.PublishOptions{
			Exchange:   rabbitmq2.Exchange_CancelEmail,
			RoutingKey: rabbitmq2.RoutingKey_Asterisk,
			Persistent: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
