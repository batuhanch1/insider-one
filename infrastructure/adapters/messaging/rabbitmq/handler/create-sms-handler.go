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

type createSmsHandler struct {
	CreateSmsCommand command.CreateCommand
}

func NewCreateSmsHandler(createSmsCommand command.CreateCommand) *createSmsHandler {
	return &createSmsHandler{createSmsCommand}
}

func (c *createSmsHandler) HandleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event sms.CreateSmsEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		err = fmt.Errorf("create sms event json unmarshal error %w", err)
		logging.Error(ctx, err)
		return err
	}

	return c.CreateSmsCommand.Execute(ctx, event)
}
