package sms

import (
	"context"
	"fmt"
	"insider-one/domain/notification/sms"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"
)

type EnqueueCancelCommand interface {
	Execute(ctx context.Context, request CancelSmsRequest) error
}

type enqueueCancelCommand struct {
	SmsRepository sms.QueryRepository
	Publisher     rabbitmq.Publisher
}

func NewEnqueueCancelCommand(smsRepository sms.QueryRepository, publisher rabbitmq.Publisher) EnqueueCancelCommand {
	return &enqueueCancelCommand{smsRepository, publisher}
}

func (s *enqueueCancelCommand) Execute(ctx context.Context, request CancelSmsRequest) error {
	smsIds, err := s.SmsRepository.GetByStatus(ctx, request.Status)
	if err != nil {
		err = fmt.Errorf("error get sms by status in cancel command: %w", err)
		logging.Error(ctx, err)
		return err
	}

	for _, smsId := range smsIds {
		cancelSmsEvent := sms.CancelSmsEvent{ID: smsId}
		err = s.Publisher.Publish(ctx, cancelSmsEvent, rabbitmq.PublishOptions{
			Exchange:   rabbitmq.Exchange_CancelSms,
			RoutingKey: rabbitmq.RoutingKey_Asterisk,
			Persistent: true,
		})
		if err != nil {
			err = fmt.Errorf("error publishing cancel sms event in cancel command : %w", err)
			logging.Error(ctx, err)
			return err
		}
	}

	return nil
}
