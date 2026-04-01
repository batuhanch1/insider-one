package sms

import (
	"context"
	"fmt"
	"insider-one/domain/notification/sms"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"

	"github.com/cespare/xxhash/v2"
)

type SendCommand interface {
	Execute(ctx context.Context, request SendSmsRequest) error
}

type sendCommand struct {
	Publisher *rabbitmq2.Publisher
}

func NewSendCommand(publisher *rabbitmq2.Publisher) SendCommand {
	return &sendCommand{publisher}
}

func (s *sendCommand) Execute(ctx context.Context, request SendSmsRequest) error {
	idempotencyString := request.PhoneNumber + request.Type + request.Sender + request.Content

	smsEvent := sms.CreateSmsEvent{
		PhoneNumber:    request.PhoneNumber,
		Sender:         request.Sender,
		Type:           request.Type,
		Content:        request.Content,
		IdempotencyKey: xxhash.Sum64String(idempotencyString),
		Priority:       request.Priority,
	}

	if request.ScheduledAt != nil {
		smsEvent.ScheduledAt = request.ScheduledAt.Unix()
	}

	err := s.Publisher.Publish(ctx, smsEvent, rabbitmq2.PublishOptions{
		Exchange:   rabbitmq2.Exchange_CreateSms,
		RoutingKey: request.Priority,
		Persistent: true,
	})
	if err != nil {
		err = fmt.Errorf("error publishing create sms event in send command: %w", err)
		logging.Error(ctx, err)
		return err
	}

	return nil
}
