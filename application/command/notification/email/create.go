package email

import (
	"context"
	"fmt"
	"insider-one/domain/notification/email"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"
)

type CreateCommand interface {
	Execute(ctx context.Context, event email.CreateEmailEvent) error
}

type createCommand struct {
	Repository email.Repository
	Publisher  *rabbitmq.Publisher
}

func NewCreateCommand(repository email.Repository, publisher *rabbitmq.Publisher) CreateCommand {
	return &createCommand{repository, publisher}
}

func (s *createCommand) Execute(ctx context.Context, event email.CreateEmailEvent) error {
	e := email.Email{
		ScheduledAt:    event.ScheduledAt,
		To:             event.To,
		From:           event.From,
		Subject:        event.Subject,
		Content:        event.Content,
		Type:           event.Type,
		IdempotencyKey: event.IdempotencyKey,
		Priority:       event.Priority,
	}
	e.SetStatus()
	err := s.Repository.Save(ctx, e)
	if err != nil {
		err = fmt.Errorf("error save email in create command: %w", err)
		logging.Error(ctx, err)
		return err
	}

	if e.IsScheduled() {
		logging.Info(ctx, "scheduled email ignored.")
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
	err = s.Publisher.Publish(ctx, emailCreatedEvent, rabbitmq.PublishOptions{
		Exchange:   rabbitmq.Exchange_EmailCreated,
		RoutingKey: event.Priority,
		Persistent: true,
	})
	if err != nil {
		err = fmt.Errorf("error publishing email created event in create command: %w", err)
		logging.Error(ctx, err)
	}
	return err
}
