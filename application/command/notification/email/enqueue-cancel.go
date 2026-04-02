package email

import (
	"context"
	"fmt"
	"insider-one/domain/notification/email"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"
)

type EnqueueCancelCommand interface {
	Execute(ctx context.Context, request CancelEmailRequest) error
}

type enqueueCancelCommand struct {
	EmailRepository email.QueryRepository
	Publisher       rabbitmq.Publisher
}

func NewEnqueueCancelCommand(emailRepository email.QueryRepository, publisher rabbitmq.Publisher) EnqueueCancelCommand {
	return &enqueueCancelCommand{emailRepository, publisher}
}

func (s *enqueueCancelCommand) Execute(ctx context.Context, request CancelEmailRequest) error {
	emailIds, err := s.EmailRepository.GetByStatus(ctx, request.Status)
	if err != nil {
		err = fmt.Errorf("error get email by status in cancel command: %w", err)
		logging.Error(ctx, err)
		return err
	}

	for _, emailId := range emailIds {
		cancelEmailEvent := email.CancelEmailEvent{ID: emailId}
		err = s.Publisher.Publish(ctx, cancelEmailEvent, rabbitmq.PublishOptions{
			Exchange:   rabbitmq.Exchange_CancelEmail,
			RoutingKey: rabbitmq.RoutingKey_Asterisk,
			Persistent: true,
		})
		if err != nil {
			err = fmt.Errorf("error publishing cancel email event in cancel command : %w", err)
			logging.Error(ctx, err)
			return err
		}
	}

	return nil
}
