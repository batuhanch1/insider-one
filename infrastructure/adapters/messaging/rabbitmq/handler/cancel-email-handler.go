package handler

import (
	"context"
	"encoding/json"
	"fmt"
	command "insider-one/application/command/notification/email"
	"insider-one/domain/notification/email"
	"insider-one/infrastructure/logging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type cancelEmailHandler struct {
	cancelEmailCommand command.CancelCommand
}

func NewCancelEmailHandler(cancelEmailCommand command.CancelCommand) *cancelEmailHandler {
	return &cancelEmailHandler{cancelEmailCommand}
}

func (c *cancelEmailHandler) HandleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event email.CancelEmailEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		err = fmt.Errorf("cancel email event json unmarshal error %w", err)
		logging.Error(ctx, err)
		return err
	}

	return c.cancelEmailCommand.Execute(ctx, event)
}
