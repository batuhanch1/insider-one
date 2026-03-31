package sms

import (
	"context"
	"insider-one/domain/notification"
	"insider-one/domain/notification/sms"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"
	"time"
)

type CreateCommand interface {
	Execute(ctx context.Context, event sms.CreateSmsEvent) error
}

type createCommand struct {
	Repository sms.Repository
	Publisher  rabbitmq2.Publisher
}

func NewCreateCommand(repository sms.Repository, publisher rabbitmq2.Publisher) CreateCommand {
	return &createCommand{repository, publisher}
}

func (c *createCommand) Execute(ctx context.Context, event sms.CreateSmsEvent) error {
	s := sms.Sms{
		ScheduledAt:    event.ScheduledAt,
		CreatedAt:      time.Now().Unix(),
		PhoneNumber:    event.PhoneNumber,
		Sender:         event.Sender,
		Type:           event.Type,
		Status:         notification.Notification_Status_Pending,
		Content:        event.Content,
		IdempotencyKey: event.IdempotencyKey,
	}
	err := c.Repository.Save(ctx, s)
	if err != nil {
		return err
	}
	if s.IsScheduled() {
		return nil
	}

	smsCreatedEvent := sms.SmsCreatedEvent{
		ScheduledAt:    s.ScheduledAt,
		IdempotencyKey: s.IdempotencyKey,
		PhoneNumber:    s.PhoneNumber,
		Sender:         s.Sender,
		Type:           s.Type,
		Status:         s.Status,
		Content:        s.Content,
	}
	err = c.Publisher.Publish(ctx, smsCreatedEvent, rabbitmq2.PublishOptions{
		Exchange:   rabbitmq2.Exchange_SmsCreated,
		RoutingKey: event.Priority,
		Persistent: true,
	})
	return err

}
