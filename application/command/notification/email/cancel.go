package email

import (
	"context"
	"fmt"
	"insider-one/domain/notification/email"
	"insider-one/infrastructure/logging"
)

type CancelCommand interface {
	Execute(ctx context.Context, event email.CancelEmailEvent) error
}

type cancelCommand struct {
	EmailRepository email.CommandRepository
}

func NewCancelCommand(emailRepository email.CommandRepository) CancelCommand {
	return &cancelCommand{emailRepository}
}

func (s *cancelCommand) Execute(ctx context.Context, event email.CancelEmailEvent) error {
	err := s.EmailRepository.Cancel(ctx, event.ID)
	if err != nil {
		err = fmt.Errorf("error cancel in cancel command: %w", err)
		logging.Error(ctx, err)
		return err
	}
	return nil
}
