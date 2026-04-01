package sms

import (
	"context"
	"fmt"
	"insider-one/domain/notification/sms"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"
	"time"
)

type ScheduleCommand interface {
	Execute(ctx context.Context) error
}

type scheduleCommand struct {
	Repository sms.Repository
	Publisher  *rabbitmq.Publisher
}

func NewScheduleCommand(repository sms.Repository, publisher *rabbitmq.Publisher) ScheduleCommand {
	return &scheduleCommand{repository, publisher}
}

func (sc *scheduleCommand) Execute(ctx context.Context) error {
	scheduledSmsList, err := sc.Repository.GetScheduled(ctx, time.Now().Add(time.Minute*5).Unix())
	if err != nil {
		err = fmt.Errorf("get scheduled smsList error in scheduled smsList: %w", err)
		logging.Error(ctx, err)
		return err
	}
	for _, s := range scheduledSmsList {
		smsCreatedEvent := sms.SmsCreatedEvent{
			ScheduledAt:    s.ScheduledAt,
			IdempotencyKey: s.IdempotencyKey,
			PhoneNumber:    s.PhoneNumber,
			Sender:         s.Sender,
			Type:           s.Type,
			Status:         s.Status,
			Content:        s.Content,
		}
		err = sc.Publisher.Publish(ctx, smsCreatedEvent, rabbitmq.PublishOptions{
			Exchange:   rabbitmq.Exchange_SmsCreated,
			RoutingKey: s.Priority,
			Persistent: true,
		})
		if err != nil {
			err = fmt.Errorf("error publishing sms created event in schedule command: %w", err)
			logging.Error(ctx, err)
		}
	}

	return nil
}
