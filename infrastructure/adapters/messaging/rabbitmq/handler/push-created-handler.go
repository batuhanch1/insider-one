package handler

import (
	"context"
	"encoding/json"
	"fmt"
	command "insider-one/application/command/notification/push"
	"insider-one/domain/notification/push"
	"insider-one/infrastructure/logging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type pushCreatedHandler struct {
	pushDeliverCommand command.DeliverCommand
}

func NewPushCreatedHandler(pushDeliverCommand command.DeliverCommand) *pushCreatedHandler {
	return &pushCreatedHandler{pushDeliverCommand}
}

func (c *pushCreatedHandler) HandleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event push.PushCreatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		err = fmt.Errorf("push created event json unmarshal error %w", err)
		logging.Error(ctx, err)
		return err
	}

	return c.pushDeliverCommand.Execute(ctx, event)
}
