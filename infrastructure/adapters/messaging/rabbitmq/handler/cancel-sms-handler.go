package handler

import (
	"context"
	"encoding/json"
	"fmt"
	command "insider-one/application/command/notification/sms"
	"insider-one/domain/notification/sms"
	"insider-one/infrastructure/logging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type cancelSmsHandler struct {
	cancelSmsCommand command.CancelCommand
}

func NewCancelSmsHandler(cancelSmsCommand command.CancelCommand) *cancelSmsHandler {
	return &cancelSmsHandler{cancelSmsCommand}
}

func (c *cancelSmsHandler) HandleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event sms.CancelSmsEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		err = fmt.Errorf("cancel Sms event json unmarshal error %w", err)
		logging.Error(ctx, err)
		return err
	}

	return c.cancelSmsCommand.Execute(ctx, event)
}
