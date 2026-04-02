package push

import (
	"context"
	"fmt"
	"insider-one/domain/notification/push"
	"insider-one/infrastructure/logging"
)

type CancelCommand interface {
	Execute(ctx context.Context, event push.CancelPushEvent) error
}

type cancelCommand struct {
	PushRepository push.CommandRepository
}

func NewCancelCommand(PushRepository push.CommandRepository) CancelCommand {
	return &cancelCommand{PushRepository}
}

func (s *cancelCommand) Execute(ctx context.Context, event push.CancelPushEvent) error {
	err := s.PushRepository.Cancel(ctx, event.ID)
	if err != nil {
		err = fmt.Errorf("error cancel in cancel command: %w", err)
		logging.Error(ctx, err)
		return err
	}
	return nil
}
