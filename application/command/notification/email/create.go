package email

import (
	"context"
	"insider-one/domain/notification"
	"insider-one/domain/notification/email"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"
)

type CreateCommand interface {
	Execute(ctx context.Context, event email.CreateEmailEvent) error
}

type createCommand struct {
	Repository email.Repository
	Publisher  rabbitmq2.Publisher
}

func NewCreateCommand(repository email.Repository, publisher rabbitmq2.Publisher) CreateCommand {
	return &createCommand{repository, publisher}
}

func (s *createCommand) Execute(ctx context.Context, event email.CreateEmailEvent) error {
	e := email.Email{
		ScheduledAt:    event.ScheduledAt,
		To:             event.To,
		From:           event.From,
		Subject:        event.Subject,
		Content:        event.Content,
		Status:         notification.Notification_Status_Pending,
		Type:           event.Type,
		IdempotencyKey: event.IdempotencyKey,
	}
	err := s.Repository.Save(ctx, e)
	if err != nil {
		return err
	}

	if e.IsScheduled() {
		return nil
	}

	emailCreatedEvent := email.EmailCreatedEvent{
		IdempotencyKey: e.IdempotencyKey,
		To:             e.To,
		From:           e.From,
		Subject:        e.Subject,
		Content:        e.Content,
		Status:         e.Status,
		Type:           e.Type,
	}
	err = s.Publisher.Publish(ctx, emailCreatedEvent, rabbitmq2.PublishOptions{
		Exchange:   rabbitmq2.Exchange_EmailCreated,
		RoutingKey: event.Priority,
		Persistent: true,
	})
	return err
}
