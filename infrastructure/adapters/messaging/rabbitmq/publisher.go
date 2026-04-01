package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"insider-one/infrastructure/logging"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(ctx context.Context, v any, opts PublishOptions) error
}

type publisher struct {
	client *Client
}

func NewPublisher(client *Client) Publisher {
	return &publisher{client: client}
}

type PublishOptions struct {
	Exchange   string
	RoutingKey string
	Headers    amqp.Table
	Persistent bool
}

func (p *publisher) Publish(ctx context.Context, v any, opts PublishOptions) error {
	body, err := json.Marshal(v)
	if err != nil {
		err = fmt.Errorf("publisher: marshal: %w", err)
		logging.Error(ctx, err)
		return err
	}

	ch, err := p.client.Channel(ctx)
	if err != nil {
		err = fmt.Errorf("publisher: open channel: %w", err)
		logging.Error(ctx, err)
		return err
	}
	defer ch.Close()

	deliveryMode := amqp.Transient
	if opts.Persistent {
		deliveryMode = amqp.Persistent
	}

	err = ch.PublishWithContext(
		ctx,
		opts.Exchange,
		opts.RoutingKey,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: deliveryMode,
			Timestamp:    time.Now(),
			Headers:      opts.Headers,
			Body:         body,
		},
	)

	if err != nil {
		err = fmt.Errorf("publisher: publish json: %w", err)
		logging.Error(ctx, err)
		return err
	}
	logging.WriteMessageToQueue(ctx, body, opts.Exchange, opts.RoutingKey)

	return nil
}
