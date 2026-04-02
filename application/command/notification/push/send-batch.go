package push

import (
	"context"
	"fmt"
	"insider-one/domain/notification"
	"insider-one/domain/notification/push"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"

	"github.com/cespare/xxhash/v2"
)

type SendBatchCommand interface {
	Execute(ctx context.Context, request SendBatchPushRequest) error
}
type sendBatchCommand struct {
	BatchPublisher rabbitmq.BatchPublisher
}

func NewSendBatchCommand(batchPublisher rabbitmq.BatchPublisher) SendBatchCommand {
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
			Priority:       request.Priority,
		}

		if !request.ScheduledAt.IsZero() {
			unix := request.ScheduledAt.Unix()
			pushEvent.ScheduledAt = &unix
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

	if err := s.publishBatch(ctx, highEventList, rabbitmq.RoutingKey_High); err != nil {
		err = fmt.Errorf("error publishing create high push event in send batch command: %w", err)
		logging.Error(ctx, err)
		return err
	}
	if err := s.publishBatch(ctx, mediumEventList, rabbitmq.RoutingKey_Medium); err != nil {
		err = fmt.Errorf("error publishing create medium push event in send batch command: %w", err)
		logging.Error(ctx, err)
		return err
	}
	if err := s.publishBatch(ctx, lowEventList, rabbitmq.RoutingKey_Low); err != nil {
		err = fmt.Errorf("error publishing create low push event in send batch command: %w", err)
		logging.Error(ctx, err)
		return err
	}

	return nil
}

func (s *sendBatchCommand) publishBatch(ctx context.Context, eventList []push.CreatePushEvent, routingKey string) error {
	if len(eventList) == 0 {
		return nil
	}
	result := make([]any, 0, len(eventList))
	for _, v := range eventList {
		result = append(result, v)
	}

	err := s.BatchPublisher.Publish(ctx, result, rabbitmq.BatchPublisherOptions{
		Exchange:   rabbitmq.Exchange_CreatePush,
		RoutingKey: routingKey,
		Persistent: true,
	})
	return err
}
