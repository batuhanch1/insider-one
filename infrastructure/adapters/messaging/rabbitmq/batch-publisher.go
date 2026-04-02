package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"insider-one/infrastructure/logging"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type BatchPublisher interface {
	Publish(ctx context.Context, vList []any, opts BatchPublisherOptions) error
}

type batchPublisher struct {
	client *Client
}

func NewBatchPublisher(client *Client) BatchPublisher {
	return &batchPublisher{client}
}

type BatchPublisherOptions struct {
	Exchange   string
	RoutingKey string
	Headers    amqp.Table
	Persistent bool
}

func (p *batchPublisher) Publish(ctx context.Context, vList []any, opts BatchPublisherOptions) error {
	if len(vList) > 1000 {
		return errors.New("maks len 1000")
	}

	channel, err := p.client.Channel(ctx)
	if err != nil {
		err = fmt.Errorf("publisher: open channel: %w", err)
		logging.Error(ctx, err)
		return err
	}
	defer channel.Close()

	if err := channel.Confirm(false); err != nil {
		return fmt.Errorf("publisher: confirm mode: %w", err)
	}

	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, len(vList)))

	publishedCount := 0
	for _, v := range vList {
		body, err := json.Marshal(v)
		if err != nil {
			logging.Error(ctx, fmt.Errorf("publisher: marshal: %w", err))
			continue
		}

		deliveryMode := amqp.Transient
		if opts.Persistent {
			deliveryMode = amqp.Persistent
		}

		err = channel.PublishWithContext(
			ctx,
			opts.Exchange,
			opts.RoutingKey,
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: deliveryMode,
				Timestamp:    time.Now(),
				Headers:      opts.Headers,
				Body:         body,
			},
		)

		if err != nil {
			logging.Error(ctx, fmt.Errorf("publisher: publish json: %w", err))
		}
		publishedCount++
	}
	for i := 0; i < publishedCount; i++ {
		select {
		case confirm, ok := <-confirms:
			if !ok {
				return errors.New("publisher: confirms channel closed unexpectedly")
			}
			if !confirm.Ack {
				logging.Error(ctx, fmt.Errorf("publisher: message NOT delivered: deliveryTag=%d", confirm.DeliveryTag))
			}
		case <-ctx.Done():
			return fmt.Errorf("publisher: context cancelled while waiting for confirms: %w", ctx.Err())
		}
	}

	return nil
}
