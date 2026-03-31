package rabbitmq

import (
	"fmt"

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

func DeclareTopology(client *Client, opts TopologyOptions) error {
	defaultTopologyOptions(&opts)

	ch, err := client.Channel()
	if err != nil {
		return fmt.Errorf("declare: open channel: %w", err)
	}
	defer ch.Close()

	if err = declareExchange(ch, opts); err != nil {
		return err
	}
	if err = declareMainQueue(ch, opts); err != nil {
		return err
	}
	if err = declareRetryQueue(ch, opts); err != nil {
		return err
	}
	if err = declareErrorQueue(ch, opts); err != nil {
		return err
	}
	if err = ch.QueueBind(opts.QueueName, opts.RoutingKey, opts.ExchangeName, false, nil); err != nil {
		return fmt.Errorf("declare: queue bind %q: %w", opts.QueueName, err)
	}

	return nil
}

func declareExchange(ch *amqp.Channel, opts TopologyOptions) error {
	if err := ch.ExchangeDeclare(
		opts.ExchangeName,
		opts.ExchangeType,
		false,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare: exchange %q: %w", opts.ExchangeName, err)
	}
	return nil
}

func declareMainQueue(ch *amqp.Channel, opts TopologyOptions) error {
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
		return fmt.Errorf("declare: main queue %q: %w", opts.QueueName, err)
	}
	return nil
}

func declareRetryQueue(ch *amqp.Channel, opts TopologyOptions) error {
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
		return fmt.Errorf("declare: retry queue %q: %w", retryQueue, err)
	}
	return nil
}

func declareErrorQueue(ch *amqp.Channel, opts TopologyOptions) error {
	errQueue := fmt.Sprintf("%s_error", opts.QueueName)
	_, err := ch.QueueDeclare(errQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare: error queue %q: %w", errQueue, err)
	}
	return nil
}
