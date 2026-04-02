package email

import (
	"context"
	"fmt"
	"insider-one/domain/notification/email"
	provider "insider-one/infrastructure/adapters/client/email-provider"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"
)

type DeliverCommand interface {
	Execute(ctx context.Context, event email.EmailCreatedEvent) error
}

type deliverCommand struct {
	Repository    email.CommandRepository
	EmailProvider provider.EmailProvider
	Publisher     rabbitmq.Publisher
}

func NewDeliverCommand(repository email.CommandRepository, emailProvider provider.EmailProvider, publisher rabbitmq.Publisher) DeliverCommand {
	return &deliverCommand{repository, emailProvider, publisher}
}

func (d *deliverCommand) Execute(ctx context.Context, event email.EmailCreatedEvent) error {
	deliverRequest := provider.NewDeliverRequest(event.To, event.From, event.Subject, event.Content)

	deliverResponse, err := d.EmailProvider.Deliver(ctx, deliverRequest)
	if err != nil {
		err = fmt.Errorf("error delivering email via provider in deliver command: %w", err)
		logging.Error(ctx, err)
		return err
	}

	if deliverResponse.IsAccepted() {
		err = d.Repository.Deliver(ctx, deliverResponse.MessageID, event.IdempotencyKey)
		if err != nil {
			err = fmt.Errorf("error delivering email on postgres in deliver command: %w", err)
			logging.Error(ctx, err)
			return err
		}
	}

	return nil
}
