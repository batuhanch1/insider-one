package rabbitmq

import (
	"context"
	"fmt"
	"insider-one/infrastructure/logging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TopologyOptions struct {
	ExchangeName string
	ExchangeType string // "fanout", "direct", "topic"
	QueueName    string
	RoutingKey   string
	RetryTTLMs   int32
}

func defaultTopologyOptions(o *TopologyOptions) {
	if o.ExchangeType == "" {
		o.ExchangeType = "fanout"
	}
	if o.RetryTTLMs == 0 {
		o.RetryTTLMs = 60_000
	}
}

func DeclareTopology(ctx context.Context, client *Client, opts TopologyOptions) error {
	defaultTopologyOptions(&opts)

	ch, err := client.Channel(ctx)
	if err != nil {
		err = fmt.Errorf("declare: open channel: %w", err)
		logging.Error(ctx, err)
		return err
	}
	defer ch.Close()

	if err = declareExchange(ctx, ch, opts); err != nil {
		return err
	}
	if err = declareMainQueue(ctx, ch, opts); err != nil {
		return err
	}
	if err = declareRetryQueue(ctx, ch, opts); err != nil {
		return err
	}
	if err = declareErrorQueue(ctx, ch, opts); err != nil {
		return err
	}
	if err = ch.QueueBind(opts.QueueName, opts.RoutingKey, opts.ExchangeName, false, nil); err != nil {
		err = fmt.Errorf("declare: queue bind %q: %w", opts.QueueName, err)
		logging.Error(ctx, err)
		return err
	}

	return nil
}

func declareExchange(ctx context.Context, ch *amqp.Channel, opts TopologyOptions) error {
	if err := ch.ExchangeDeclare(
		opts.ExchangeName,
		opts.ExchangeType,
		false,
		false,
		false,
		false,
		nil,
	); err != nil {
		err = fmt.Errorf("declare: exchange %q: %w", opts.ExchangeName, err)
		logging.Error(ctx, err)
		return err
	}
	return nil
}

func declareMainQueue(ctx context.Context, ch *amqp.Channel, opts TopologyOptions) error {
	retryQueue := fmt.Sprintf("%s_retry", opts.QueueName)
	_, err := ch.QueueDeclare(
		opts.QueueName,
		true,
		false,
		false,
		false,
		amqp.Table{
			amqp.QueueTypeArg:           amqp.QueueTypeQuorum,
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": retryQueue,
		},
	)
	if err != nil {
		err = fmt.Errorf("declare: main queue %q: %w", opts.QueueName, err)
		logging.Error(ctx, err)
		return err
	}
	return nil
}

func declareRetryQueue(ctx context.Context, ch *amqp.Channel, opts TopologyOptions) error {
	retryQueue := fmt.Sprintf("%s_retry", opts.QueueName)
	_, err := ch.QueueDeclare(
		retryQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			amqp.QueueTypeArg:           amqp.QueueTypeQuorum,
			"x-message-ttl":             opts.RetryTTLMs,
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": opts.QueueName,
		},
	)
	if err != nil {
		err = fmt.Errorf("declare: retry queue %q: %w", retryQueue, err)
		logging.Error(ctx, err)
		return err
	}
	return nil
}

func declareErrorQueue(ctx context.Context, ch *amqp.Channel, opts TopologyOptions) error {
	errQueue := fmt.Sprintf("%s_error", opts.QueueName)
	_, err := ch.QueueDeclare(errQueue, true, false, false, false, nil)
	if err != nil {
		err = fmt.Errorf("declare: error queue %q: %w", errQueue, err)
		logging.Error(ctx, err)
		return err
	}
	return nil
}
