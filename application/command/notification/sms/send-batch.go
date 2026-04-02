package sms

import (
	"context"
	"fmt"
	"insider-one/domain/notification"
	"insider-one/domain/notification/sms"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"

	"github.com/cespare/xxhash/v2"
)

type SendBatchCommand interface {
	Execute(ctx context.Context, request SendBatchSmsRequest) error
}
type sendBatchCommand struct {
	BatchPublisher rabbitmq.BatchPublisher
}

func NewSendBatchCommand(batchPublisher rabbitmq.BatchPublisher) SendBatchCommand {
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
			Priority:       request.Priority,
		}

		if !request.ScheduledAt.IsZero() {
			unix := request.ScheduledAt.Unix()
			smsEvent.ScheduledAt = &unix
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

	if err := s.publishBatch(ctx, highEventList, rabbitmq.RoutingKey_High); err != nil {
		err = fmt.Errorf("error publishing create high sms event in send batch command: %w", err)
		logging.Error(ctx, err)
		return err
	}
	if err := s.publishBatch(ctx, mediumEventList, rabbitmq.RoutingKey_Medium); err != nil {
		err = fmt.Errorf("error publishing create medium sms event in send batch command: %w", err)
		logging.Error(ctx, err)
		return err
	}
	if err := s.publishBatch(ctx, lowEventList, rabbitmq.RoutingKey_Low); err != nil {
		err = fmt.Errorf("error publishing create low sms event in send batch command: %w", err)
		logging.Error(ctx, err)
		return err
	}

	return nil
}

func (s *sendBatchCommand) publishBatch(ctx context.Context, eventList []sms.CreateSmsEvent, routingKey string) error {
	if len(eventList) == 0 {
		return nil
	}
	result := make([]any, 0, len(eventList))
	for _, v := range eventList {
		result = append(result, v)
	}

	err := s.BatchPublisher.Publish(ctx, result, rabbitmq.BatchPublisherOptions{
		Exchange:   rabbitmq.Exchange_CreateSms,
		RoutingKey: routingKey,
		Persistent: true,
	})
	return err
}
