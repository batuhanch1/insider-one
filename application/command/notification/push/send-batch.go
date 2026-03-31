package push

import (
	"context"
	"insider-one/domain/notification"
	"insider-one/domain/notification/push"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/cespare/xxhash/v2"
)

type SendBatchCommand interface {
	Execute(ctx context.Context, request SendBatchPushRequest) error
}
type sendBatchCommand struct {
	BatchPublisher *rabbitmq2.BatchPublisher
}

func NewSendBatchCommand(batchPublisher *rabbitmq2.BatchPublisher) SendBatchCommand {
	return &sendBatchCommand{batchPublisher}
}

func (s *sendBatchCommand) Execute(ctx context.Context, batchRequest SendBatchPushRequest) error {
	var highEventList []push.CreatePushEvent
	var mediumEventList []push.CreatePushEvent
	var lowEventList []push.CreatePushEvent

	for _, request := range batchRequest.Pushes {
		idempotencyString := request.PhoneNumber + request.Type + request.Sender + request.Content
		pushEvent := push.CreatePushEvent{
			Sender:         request.Sender,
			PhoneNumber:    request.PhoneNumber,
			Type:           request.Type,
			Content:        request.Content,
			IdempotencyKey: xxhash.Sum64String(idempotencyString),
		}

		if request.ScheduledAt != nil {
			pushEvent.ScheduledAt = request.ScheduledAt.Unix()
		}

		switch request.Priority {
		case notification.Notification_Priority_High:
			highEventList = append(highEventList, pushEvent)
		case notification.Notification_Priority_Medium:
			mediumEventList = append(mediumEventList, pushEvent)
		case notification.Notification_Priority_Low:
			lowEventList = append(lowEventList, pushEvent)
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

func (s *sendBatchCommand) publishBatch(ctx context.Context, eventList []push.CreatePushEvent, routingKey string) error {
	result := make([]any, 0, len(eventList))
	for _, v := range eventList {
		result = append(result, v)
	}

	err := s.BatchPublisher.Publish(ctx, result, rabbitmq2.BatchPublisherOptions{
		Exchange:   rabbitmq2.Exchange_CreatePush,
		RoutingKey: routingKey,
		Persistent: true,
	})
	return err
}
