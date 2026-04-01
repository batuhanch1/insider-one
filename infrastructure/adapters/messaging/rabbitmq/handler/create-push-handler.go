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

type createPushHandler struct {
	CreatePushCommand command.CreateCommand
}

func NewCreatePushHandler(createPushCommand command.CreateCommand) *createPushHandler {
	return &createPushHandler{createPushCommand}
}

func (c *createPushHandler) HandleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event push.CreatePushEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		err = fmt.Errorf("create push event json unmarshal error %w", err)
		logging.Error(ctx, err)
		return err
	}

	return c.CreatePushCommand.Execute(ctx, event)
}
