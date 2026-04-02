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

type cancelPushHandler struct {
	cancelPushCommand command.CancelCommand
}

func NewCancelPushHandler(cancelPushCommand command.CancelCommand) *cancelPushHandler {
	return &cancelPushHandler{cancelPushCommand}
}

func (c *cancelPushHandler) HandleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event push.CancelPushEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		err = fmt.Errorf("cancel Push event json unmarshal error %w", err)
		logging.Error(ctx, err)
		return err
	}

	return c.cancelPushCommand.Execute(ctx, event)
}
