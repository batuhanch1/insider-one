package sms

import (
	"context"
	"fmt"
	"insider-one/domain/notification/sms"
	provider "insider-one/infrastructure/adapters/client/sms-provider"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"
)

type DeliverCommand interface {
	Execute(ctx context.Context, event sms.SmsCreatedEvent) error
}

type deliverCommand struct {
	Repository  sms.Repository
	smsProvider provider.SmsProvider
	Publisher   rabbitmq.Publisher
}

func NewDeliverCommand(repository sms.Repository, smsProvider provider.SmsProvider, publisher rabbitmq.Publisher) DeliverCommand {
	return &deliverCommand{repository, smsProvider, publisher}
}

func (d *deliverCommand) Execute(ctx context.Context, event sms.SmsCreatedEvent) error {
	deliverRequest := provider.NewDeliverRequest(event.Sender, event.PhoneNumber, event.Content)

	deliverResponse, err := d.smsProvider.Deliver(ctx, deliverRequest)
	if err != nil {
		err = fmt.Errorf("error delivering sms via provider in deliver command: %w", err)
		logging.Error(ctx, err)
		return err
	}

	if deliverResponse.IsAccepted() {
		err = d.Repository.Deliver(ctx, deliverResponse.MessageID, event.IdempotencyKey)
		if err != nil {
			err = fmt.Errorf("error delivering sms on postgres in deliver command: %w", err)
			logging.Error(ctx, err)
			return err
		}
	}

	return nil
}
