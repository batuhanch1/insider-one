package email

import (
	"context"
	"fmt"
	"insider-one/domain/notification/email"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"

	"github.com/cespare/xxhash/v2"
)

type SendCommand interface {
	Execute(ctx context.Context, request SendEmailRequest) error
}

type sendCommand struct {
	Publisher *rabbitmq2.Publisher
}

func NewSendCommand(publisher *rabbitmq2.Publisher) SendCommand {
	return &sendCommand{publisher}
}

func (s *sendCommand) Execute(ctx context.Context, request SendEmailRequest) error {
	idempotencyString := request.From + request.To + request.Subject + request.Content
	emailEvent := email.CreateEmailEvent{
		To:             request.To,
		From:           request.From,
		Subject:        request.Subject,
		Content:        request.Content,
		Type:           request.Type,
		IdempotencyKey: xxhash.Sum64String(idempotencyString),
		Priority:       request.Priority,
	}

	if request.ScheduledAt != nil {
		emailEvent.ScheduledAt = request.ScheduledAt.Unix()
	}

	err := s.Publisher.Publish(ctx, emailEvent, rabbitmq2.PublishOptions{
		Exchange:   rabbitmq2.Exchange_CreateEmail,
		RoutingKey: request.Priority,
		Persistent: true,
	})
	if err != nil {
		err = fmt.Errorf("error publishing create email event in send command: %w", err)
		logging.Error(ctx, err)
		return err
	}

	return nil
}
