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

type emailCreatedHandler struct {
	emailDeliverCommand command.DeliverCommand
}

func NewEmailCreatedHandler(emailDeliverCommand command.DeliverCommand) *emailCreatedHandler {
	return &emailCreatedHandler{emailDeliverCommand}
}

func (c *emailCreatedHandler) HandleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event email.EmailCreatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		err = fmt.Errorf("email created event json unmarshal error %w", err)
		logging.Error(ctx, err)
		return err
	}

	return c.emailDeliverCommand.Execute(ctx, event)
}
