package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"insider-one/infrastructure/logging"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type BatchPublisher interface {
	Publish(ctx context.Context, vList []any, opts BatchPublisherOptions) error
}

type batchPublisher struct {
	channel *amqp.Channel
}

func NewBatchPublisher(channel *amqp.Channel) BatchPublisher {
	return &batchPublisher{channel}
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

	p.channel.Confirm(false)
	confirms := p.channel.NotifyPublish(make(chan amqp.Confirmation, 1000))

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		for confirm := range confirms {
			if !confirm.Ack {
				logging.Error(ctx, errors.New(fmt.Sprintf("Message NOT delivered: %d", confirm.DeliveryTag)))
			}
		}
	}()

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

		err = p.channel.PublishWithContext(
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
	}
	wg.Wait()
	return nil
}
