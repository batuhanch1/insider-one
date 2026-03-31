package push

import (
	"context"
	"insider-one/domain/notification/push"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/cespare/xxhash/v2"
)

type SendCommand interface {
	Execute(ctx context.Context, request SendPushRequest) error
}

type sendCommand struct {
	publisher *rabbitmq2.Publisher
}

func NewSendCommand(publisher *rabbitmq2.Publisher) SendCommand {
	return &sendCommand{publisher}
}

func (s *sendCommand) Execute(ctx context.Context, request SendPushRequest) error {
	idempotencyString := request.PhoneNumber + request.Type + request.Sender + request.Content
	pushEvent := push.CreatePushEvent{
		Sender:         request.Sender,
		PhoneNumber:    request.PhoneNumber,
		Type:           request.Type,
		Content:        request.Content,
		IdempotencyKey: xxhash.Sum64String(idempotencyString),
		Priority:       request.Priority,
	}

	if request.ScheduledAt != nil {
		pushEvent.ScheduledAt = request.ScheduledAt.Unix()
	}

	err := s.publisher.Publish(ctx, pushEvent, rabbitmq2.PublishOptions{
		Exchange:   rabbitmq2.Exchange_CreatePush,
		RoutingKey: request.Priority,
		Persistent: true,
	})
	if err != nil {
		return err
	}

	return nil
}
