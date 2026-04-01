package push

import (
	"context"
	"fmt"
	"insider-one/domain/notification/push"
	provider "insider-one/infrastructure/adapters/client/Push-provider"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/logging"
)

type DeliverCommand interface {
	Execute(ctx context.Context, event push.PushCreatedEvent) error
}

type deliverCommand struct {
	Repository   push.Repository
	PushProvider provider.PushProvider
	Publisher    *rabbitmq.Publisher
}

func NewDeliverCommand(repository push.Repository, PushProvider provider.PushProvider, publisher *rabbitmq.Publisher) DeliverCommand {
	return &deliverCommand{repository, PushProvider, publisher}
}

func (d *deliverCommand) Execute(ctx context.Context, event push.PushCreatedEvent) error {
	deliverRequest := provider.NewDeliverRequest(event.Sender, event.PhoneNumber, event.Content)

	deliverResponse, err := d.PushProvider.Deliver(ctx, deliverRequest)
	if err != nil {
		err = fmt.Errorf("error delivering push via provider in deliver command: %w", err)
		logging.Error(ctx, err)
		return err
	}

	if deliverResponse.IsAccepted() {
		err = d.Repository.Deliver(ctx, deliverResponse.MessageID, event.IdempotencyKey)
		if err != nil {
			err = fmt.Errorf("error delivering push on postgres in deliver command: %w", err)
			logging.Error(ctx, err)
			return err
		}
	}

	return nil
}
