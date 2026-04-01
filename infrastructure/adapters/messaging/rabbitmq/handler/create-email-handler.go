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

type createEmailHandler struct {
	CreateEmailCommand command.CreateCommand
}

func NewCreateEmailHandler(createEmailCommand command.CreateCommand) *createEmailHandler {
	return &createEmailHandler{createEmailCommand}
}

func (c *createEmailHandler) HandleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event email.CreateEmailEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		err = fmt.Errorf("create email event json unmarshal error %w", err)
		logging.Error(ctx, err)
		return err
	}

	return c.CreateEmailCommand.Execute(ctx, event)
}
