package email

import (
	"context"
	"insider-one/domain/notification"
	"insider-one/domain/notification/email"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/cespare/xxhash/v2"
)

type SendBatchCommand interface {
	Execute(ctx context.Context, request SendBatchEmailRequest) error
}
type sendBatchCommand struct {
	BatchPublisher *rabbitmq2.BatchPublisher
}

func NewSendBatchCommand(batchPublisher *rabbitmq2.BatchPublisher) SendBatchCommand {
	return &sendBatchCommand{batchPublisher}
}

func (s *sendBatchCommand) Execute(ctx context.Context, batchRequest SendBatchEmailRequest) error {
	var highEventList []email.CreateEmailEvent
	var mediumEventList []email.CreateEmailEvent
	var lowEventList []email.CreateEmailEvent
	for _, request := range batchRequest.Emails {
		idempotencyString := request.From + request.To + request.Type + request.Subject + request.Content
		emailEvent := email.CreateEmailEvent{
			To:             request.To,
			From:           request.From,
			Subject:        request.Subject,
			Content:        request.Content,
			Type:           request.Type,
			IdempotencyKey: xxhash.Sum64String(idempotencyString),
		}
		if request.ScheduledAt != nil {
			emailEvent.ScheduledAt = request.ScheduledAt.Unix()
		}
		switch request.Priority {
		case notification.Notification_Priority_High:
			highEventList = append(highEventList, emailEvent)
		case notification.Notification_Priority_Medium:
			mediumEventList = append(mediumEventList, emailEvent)
		case notification.Notification_Priority_Low:
			lowEventList = append(lowEventList, emailEvent)
		}
	}
	if err := s.publishBatch(ctx, highEventList, rabbitmq2.RoutingKey_High); err != nil {
		return err
	}
	if err := s.publishBatch(ctx, mediumEventList, rabbitmq2.RoutingKey_Medium); err != nil {
		return err
	}
	if err := s.publishBatch(ctx, lowEventList, rabbitmq2.RoutingKey_Low); err != nil {
		return err
	}

	return nil
}
func (s *sendBatchCommand) publishBatch(ctx context.Context, eventList []email.CreateEmailEvent, routingKey string) error {
	result := make([]any, 0, len(eventList))
	for _, v := range eventList {
		result = append(result, v)
	}

	err := s.BatchPublisher.Publish(ctx, result, rabbitmq2.BatchPublisherOptions{
		Exchange:   rabbitmq2.Exchange_CreateEmail,
		RoutingKey: routingKey,
		Persistent: true,
	})
	return err
}
