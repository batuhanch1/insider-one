package sms

import (
	"context"
	"fmt"
	"insider-one/domain/notification/sms"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"
)

type CreateCommand interface {
	Execute(ctx context.Context, event sms.CreateSmsEvent) error
}

type createCommand struct {
	Repository sms.Repository
	Publisher  rabbitmq.Publisher
}

func NewCreateCommand(repository sms.Repository, publisher rabbitmq.Publisher) CreateCommand {
	return &createCommand{repository, publisher}
}

func (c *createCommand) Execute(ctx context.Context, event sms.CreateSmsEvent) error {
	s := sms.Sms{
		ScheduledAt:    event.ScheduledAt,
		PhoneNumber:    event.PhoneNumber,
		Sender:         event.Sender,
		Type:           event.Type,
		Content:        event.Content,
		IdempotencyKey: event.IdempotencyKey,
		Priority:       event.Priority,
	}
	s.SetStatus()
	err := c.Repository.Save(ctx, s)
	if err != nil {
		err = fmt.Errorf("error save sms in create command: %w", err)
		logging.Error(ctx, err)
		return err
	}
	if s.IsScheduled() {
		logging.Info(ctx, "scheduled push ignored.")
		return nil
	}

	smsCreatedEvent := sms.SmsCreatedEvent{
		IdempotencyKey: s.IdempotencyKey,
		PhoneNumber:    s.PhoneNumber,
		Sender:         s.Sender,
		Type:           s.Type,
		Status:         s.Status,
		Content:        s.Content,
	}
	err = c.Publisher.Publish(ctx, smsCreatedEvent, rabbitmq.PublishOptions{
		Exchange:   rabbitmq.Exchange_SmsCreated,
		RoutingKey: event.Priority,
		Persistent: true,
	})
	if err != nil {
		err = fmt.Errorf("error publishing sms created event in create command: %w", err)
		logging.Error(ctx, err)
	}

	return err

}
