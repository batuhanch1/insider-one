package sms

import (
	"context"
	"fmt"
	"insider-one/domain/notification/sms"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"
)

type CancelCommand interface {
	Execute(ctx context.Context, request CancelSmsRequest) error
}

type cancelCommand struct {
	SmsRepository sms.Repository
	Publisher     *rabbitmq2.Publisher
}

func NewCancelCommand(smsRepository sms.Repository, publisher *rabbitmq2.Publisher) CancelCommand {
	return &cancelCommand{smsRepository, publisher}
}

func (s *cancelCommand) Execute(ctx context.Context, request CancelSmsRequest) error {
	smsIds, err := s.SmsRepository.GetByStatus(ctx, request.Status)
	if err != nil {
		err = fmt.Errorf("error get sms by status in cancel command: %w", err)
		logging.Error(ctx, err)
		return err
	}

	for _, smsId := range smsIds {
		cancelSmsEvent := sms.CancelSmsEvent{ID: smsId}
		err = s.Publisher.Publish(ctx, cancelSmsEvent, rabbitmq2.PublishOptions{
			Exchange:   rabbitmq2.Exchange_CancelSms,
			RoutingKey: rabbitmq2.RoutingKey_Asterisk,
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
