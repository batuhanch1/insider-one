package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"insider-one/infrastructure/logging"
	prometheus2 "insider-one/infrastructure/prometheus"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/time/rate"
)

type MessageHandler func(ctx context.Context, msg amqp.Delivery) error

const rateLimit = 100

type Consumer struct {
	client            *Client
	queueName         string
	handler           MessageHandler
	opts              ConsumerOptions
	PrometheusWrapper *prometheus2.Prometheus
}

type ConsumerOptions struct {
	WorkerCount   int
	PrefetchCount int
	MaxRetry      int
}

func defaultConsumerOptions() ConsumerOptions {
	return ConsumerOptions{
		WorkerCount:   10,
		PrefetchCount: 10,
		MaxRetry:      3,
	}
}

func NewConsumer(client *Client, queueName string, handler MessageHandler, prometheusWrapper *prometheus2.Prometheus, optFns ...func(*ConsumerOptions)) *Consumer {
	opts := defaultConsumerOptions()
	for _, fn := range optFns {
		fn(&opts)
	}
	return &Consumer{
		client:            client,
		queueName:         queueName,
		handler:           handler,
		opts:              opts,
		PrometheusWrapper: prometheusWrapper,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		limiter := rate.NewLimiter(rate.Limit(rateLimit), 1)

		if err := c.consume(ctx, limiter); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			logging.Error(ctx, fmt.Errorf("consumer stopped unexpectedly, retrying. queue :%s, err: %w", c.queueName, err))
			select {
			case <-c.client.reconnectC:
			case <-time.After(5 * time.Second):
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func (c *Consumer) consume(ctx context.Context, limiter *rate.Limiter) error {
	ch, err := c.client.Channel(ctx)
	if err != nil {
		err = fmt.Errorf("open channel: %w", err)
		logging.Error(ctx, err)
		return err
	}
	defer ch.Close()

	if c.opts.PrefetchCount < c.opts.WorkerCount {
		err = errors.New("prefetch count must be >= worker count")
		logging.Error(ctx, err)
		return err
	}

	if err = ch.Qos(c.opts.PrefetchCount, 0, false); err != nil {
		err = fmt.Errorf("set qos: %w, queueName : %s", err, c.queueName)
		logging.Error(ctx, err)
		return err
	}

	msgs, err := ch.Consume(c.queueName, "", false, false, false, false, nil)
	if err != nil {
		err = fmt.Errorf("start consume: %w, queueName : %s", err, c.queueName)
		logging.Error(ctx, err)
		return err
	}

	jobs := make(chan amqp.Delivery, c.opts.WorkerCount)
	var wg sync.WaitGroup

	for i := 0; i < c.opts.WorkerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for msg := range jobs {
				c.processMessage(ctx, ch, msg, id)
			}
		}(i)
	}

	func() {
		defer close(jobs)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				if err = limiter.Wait(ctx); err != nil {
					msg.Nack(false, true)
					return
				}
				select {
				case jobs <- msg:
				case <-ctx.Done():
					msg.Nack(false, true)
					return
				}
			}
		}
	}()

	wg.Wait()
	return nil
}

func (c *Consumer) processMessage(ctx context.Context, ch *amqp.Channel, msg amqp.Delivery, workerID int) {
	start := time.Now()
	defer func() {
		c.PrometheusWrapper.ConsumerProcessingDuration.WithLabelValues(c.queueName).
			Observe(time.Since(start).Seconds())
	}()

	logging.ReadMessageFromQueue(ctx, msg.Body, c.queueName, msg.RoutingKey)
	err := c.handler(ctx, msg)
	if err == nil {
		c.PrometheusWrapper.ConsumerMessagesTotal.WithLabelValues(c.queueName, "success").Inc()
		msg.Ack(false)
		return
	}
	c.PrometheusWrapper.ConsumerMessagesTotal.WithLabelValues(c.queueName, "failed").Inc()

	logging.Error(ctx, fmt.Errorf("message handler failed. worker: %v, queue: %s, error: %w", workerID, c.queueName, err))
	retryOrDLQ(ctx, ch, msg, c.queueName, c.opts.MaxRetry)
}

func retryOrDLQ(ctx context.Context, ch *amqp.Channel, msg amqp.Delivery, queueName string, maxRetry int) {
	retryCount := getRetryCount(msg.Headers)
	retryQueue := fmt.Sprintf("%s_retry", queueName)
	dlqQueue := fmt.Sprintf("%s_error", queueName)

	target := dlqQueue
	if retryCount < maxRetry {
		target = retryQueue
	}

	headers := msg.Headers
	if headers == nil {
		headers = amqp.Table{}
	}
	headers["x-retry-count"] = int32(retryCount + 1)

	if err := ch.Publish("", target, false, false, amqp.Publishing{
		Body:    msg.Body,
		Headers: headers,
	}); err != nil {
		logging.Error(ctx, fmt.Errorf("failed to route message. target: %s,err :%w", target, err))
		msg.Nack(false, true)
		return
	}
	msg.Ack(false)
}

func getRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	if v, ok := headers["x-retry-count"]; ok {
		if val, ok := v.(int32); ok {
			return int(val)
		}
	}
	return 0
}
