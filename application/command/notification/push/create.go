package push

import (
	"context"
	"insider-one/domain/notification"
	"insider-one/domain/notification/push"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"
)

type CreateCommand interface {
	Execute(ctx context.Context, event push.CreatePushEvent) error
}

type createCommand struct {
	Repository push.Repository
	Publisher  rabbitmq2.Publisher
}

func NewCreateCommand(repository push.Repository, publisher rabbitmq2.Publisher) CreateCommand {
	return &createCommand{repository, publisher}
}

func (s *createCommand) Execute(ctx context.Context, event push.CreatePushEvent) error {
	p := push.Push{
		ScheduledAt:    event.ScheduledAt,
		Sender:         event.Sender,
		PhoneNumber:    event.PhoneNumber,
		Type:           event.Type,
		Status:         notification.Notification_Status_Pending,
		Content:        event.Content,
		IdempotencyKey: event.IdempotencyKey,
	}
	err := s.Repository.Save(ctx, p)
	if err != nil {
		return err
	}

	if p.IsScheduled() {
		return nil
	}

	pushCreatedEvent := push.PushCreatedEvent{
		ScheduledAt:    p.ScheduledAt,
		IdempotencyKey: p.IdempotencyKey,
		Sender:         p.Sender,
		PhoneNumber:    p.PhoneNumber,
		Type:           p.Type,
		Status:         p.Status,
		Content:        p.Content,
	}
	err = s.Publisher.Publish(ctx, pushCreatedEvent, rabbitmq2.PublishOptions{
		Exchange:   rabbitmq2.Exchange_PushCreated,
		RoutingKey: event.Priority,
		Persistent: true,
	})

	return err
}
