package email

import (
	"context"
	"fmt"
	"insider-one/domain/notification/email"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"
	"time"
)

type ScheduleCommand interface {
	Execute(ctx context.Context) error
}

type scheduleCommand struct {
	Repository email.Repository
	Publisher  rabbitmq.Publisher
}

func NewScheduleCommand(repository email.Repository, publisher rabbitmq.Publisher) ScheduleCommand {
	return &scheduleCommand{repository, publisher}
}

func (s *scheduleCommand) Execute(ctx context.Context) error {
	scheduledEmails, err := s.Repository.GetScheduled(ctx, time.Now().Add(time.Minute*5).Unix())
	if err != nil {
		err = fmt.Errorf("get scheduled emails error in scheduled emails: %w", err)
		logging.Error(ctx, err)
		return err
	}
	for _, e := range scheduledEmails {
		emailCreatedEvent := email.EmailCreatedEvent{
			ScheduledAt:    e.ScheduledAt,
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
			RoutingKey: e.Priority,
			Persistent: true,
		})
		if err != nil {
			err = fmt.Errorf("error publishing email created event in schedule command: %w", err)
			logging.Error(ctx, err)
		}
	}

	return nil
}
