package push

import (
	"context"
	"fmt"
	"insider-one/domain/notification/push"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"
)

type CancelCommand interface {
	Execute(ctx context.Context, request CancelPushRequest) error
}

type cancelCommand struct {
	PushRepository push.Repository
	Publisher      *rabbitmq.Publisher
}

func NewCancelCommand(pushRepository push.Repository, publisher *rabbitmq.Publisher) CancelCommand {
	return &cancelCommand{pushRepository, publisher}
}

func (s *cancelCommand) Execute(ctx context.Context, request CancelPushRequest) error {
	pushIds, err := s.PushRepository.GetByStatus(ctx, request.Status)
	if err != nil {
		err = fmt.Errorf("error get push by status in cancel command: %w", err)
		logging.Error(ctx, err)
		return err
	}

	for _, pushId := range pushIds {
		cancelPushEvent := push.CancelPushEvent{ID: pushId}
		err = s.Publisher.Publish(ctx, cancelPushEvent, rabbitmq.PublishOptions{
			Exchange:   rabbitmq.Exchange_CancelPush,
			RoutingKey: rabbitmq.RoutingKey_Asterisk,
			Persistent: true,
		})
		if err != nil {
			err = fmt.Errorf("error publishing cancel push event in cancel command : %w", err)
			logging.Error(ctx, err)
			return err
		}
	}

	return nil
}
