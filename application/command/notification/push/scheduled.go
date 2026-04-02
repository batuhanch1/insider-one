package push

import (
	"context"
	"fmt"
	"insider-one/domain/notification/push"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"
	"time"
)

type ScheduleCommand interface {
	Execute(ctx context.Context) error
}

type scheduleCommand struct {
	Repository push.QueryRepository
	Publisher  rabbitmq.Publisher
}

func NewScheduleCommand(repository push.QueryRepository, publisher rabbitmq.Publisher) ScheduleCommand {
	return &scheduleCommand{repository, publisher}
}

func (s *scheduleCommand) Execute(ctx context.Context) error {
	scheduledPushes, err := s.Repository.GetScheduled(ctx, time.Now().Add(time.Minute*5).Unix())
	if err != nil {
		err = fmt.Errorf("get scheduled pushes error in scheduled pushes: %w", err)
		logging.Error(ctx, err)
		return err
	}
	for _, p := range scheduledPushes {
		pushCreatedEvent := push.PushCreatedEvent{
			ScheduledAt:    p.ScheduledAt,
			IdempotencyKey: p.IdempotencyKey,
			Sender:         p.Sender,
			PhoneNumber:    p.PhoneNumber,
			Type:           p.Type,
			Status:         p.Status,
			Content:        p.Content,
		}
		err = s.Publisher.Publish(ctx, pushCreatedEvent, rabbitmq.PublishOptions{
			Exchange:   rabbitmq.Exchange_PushCreated,
			RoutingKey: p.Priority,
			Persistent: true,
		})

		if err != nil {
			err = fmt.Errorf("error publishing push created event in schedule command: %w", err)
			logging.Error(ctx, err)
		}
	}

	return nil
}
