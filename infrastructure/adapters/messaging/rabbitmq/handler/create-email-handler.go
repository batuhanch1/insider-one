package handler

import (
	"context"
	"encoding/json"
	command "insider-one/application/command/notification/email"
	"insider-one/domain/notification/email"

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
		return err
	}

	return c.CreateEmailCommand.Execute(ctx, event)
}
