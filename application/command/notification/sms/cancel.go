package sms

import (
	"context"
	"fmt"
	"insider-one/domain/notification/sms"
	"insider-one/infrastructure/logging"
)

type CancelCommand interface {
	Execute(ctx context.Context, event sms.CancelSmsEvent) error
}

type cancelCommand struct {
	SmsRepository sms.CommandRepository
}

func NewCancelCommand(SmsRepository sms.CommandRepository) CancelCommand {
	return &cancelCommand{SmsRepository}
}

func (s *cancelCommand) Execute(ctx context.Context, event sms.CancelSmsEvent) error {
	err := s.SmsRepository.Cancel(ctx, event.ID)
	if err != nil {
		err = fmt.Errorf("error cancel in cancel command: %w", err)
		logging.Error(ctx, err)
		return err
	}
	return nil
}
