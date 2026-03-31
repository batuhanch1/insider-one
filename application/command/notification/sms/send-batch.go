package sms

import (
	"context"
	"insider-one/domain/notification"
	"insider-one/domain/notification/sms"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/cespare/xxhash/v2"
)

type SendBatchCommand interface {
	Execute(ctx context.Context, request SendBatchSmsRequest) error
}
type sendBatchCommand struct {
	BatchPublisher *rabbitmq2.BatchPublisher
}

func NewSendBatchCommand(batchPublisher *rabbitmq2.BatchPublisher) SendBatchCommand {
	return &sendBatchCommand{batchPublisher}
}

func (s *sendBatchCommand) Execute(ctx context.Context, batchRequest SendBatchSmsRequest) error {
	var highEventList []sms.CreateSmsEvent
	var mediumEventList []sms.CreateSmsEvent
	var lowEventList []sms.CreateSmsEvent
	for _, request := range batchRequest.Sms {
		idempotencyString := request.PhoneNumber + request.Type + request.Sender + request.Content
		smsEvent := sms.CreateSmsEvent{
			PhoneNumber:    request.PhoneNumber,
			Sender:         request.Sender,
			Type:           request.Type,
			Content:        request.Content,
			IdempotencyKey: xxhash.Sum64String(idempotencyString),
		}

		if request.ScheduledAt != nil {
			smsEvent.ScheduledAt = request.ScheduledAt.Unix()
		}

		switch request.Priority {
		case notification.Notification_Priority_High:
			highEventList = append(highEventList, smsEvent)
		case notification.Notification_Priority_Medium:
			mediumEventList = append(mediumEventList, smsEvent)
		case notification.Notification_Priority_Low:
			lowEventList = append(lowEventList, smsEvent)
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

func (s *sendBatchCommand) publishBatch(ctx context.Context, eventList []sms.CreateSmsEvent, routingKey string) error {
	result := make([]any, 0, len(eventList))
	for _, v := range eventList {
		result = append(result, v)
	}

	err := s.BatchPublisher.Publish(ctx, result, rabbitmq2.BatchPublisherOptions{
		Exchange:   rabbitmq2.Exchange_CreateSms,
		RoutingKey: routingKey,
		Persistent: true,
	})
	return err
}
