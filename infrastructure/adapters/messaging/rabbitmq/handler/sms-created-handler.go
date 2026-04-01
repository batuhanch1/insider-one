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

type smsCreatedHandler struct {
	smsDeliverCommand command.DeliverCommand
}

func NewSmsCreatedHandler(smsDeliverCommand command.DeliverCommand) *smsCreatedHandler {
	return &smsCreatedHandler{smsDeliverCommand}
}

func (c *smsCreatedHandler) HandleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event sms.SmsCreatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		err = fmt.Errorf("sms created event json unmarshal error %w", err)
		logging.Error(ctx, err)
		return err
	}

	return c.smsDeliverCommand.Execute(ctx, event)
}
