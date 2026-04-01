package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"insider-one/infrastructure/config"
	"insider-one/infrastructure/logging"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	cfg        config.Config
	conn       *amqp.Connection
	mu         sync.RWMutex
	done       chan struct{}
	reconnectC chan struct{}
}

func New(ctx context.Context, cfg *config.Config) (*Client, error) {
	c := &Client{
		cfg:        *cfg,
		done:       make(chan struct{}),
		reconnectC: make(chan struct{}, 1),
	}
	if err := c.connect(ctx); err != nil {
		return nil, err
	}
	go c.watchConnection(ctx)
	return c, nil
}

func (c *Client) connect(ctx context.Context) error {
	url := fmt.Sprintf("AMQP://%s:%s@%s:%d/", c.cfg.Rabbit.User, c.cfg.Rabbit.Password, c.cfg.Rabbit.Host, c.cfg.Rabbit.Port)
	conn, err := amqp.Dial(url)
	if err != nil {
		err = fmt.Errorf("AMQP dial: %w", err)
		logging.Error(ctx, err)
		return err
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	return nil
}

func (c *Client) watchConnection(ctx context.Context) {
	for {
		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		errCh := conn.NotifyClose(make(chan *amqp.Error, 1))
		select {
		case amqpErr, ok := <-errCh:
			if !ok {
				return
			}
			logging.Error(ctx, fmt.Errorf("RabbitMQ connection lost, reconnecting. error: %w", amqpErr))
			c.reconnectWithBackoff(ctx)

		case <-c.done:
			return
		}
	}
}

func (c *Client) reconnectWithBackoff(ctx context.Context) {
	backoff := 1 * time.Second
	for attempt := 1; ; attempt++ {
		select {
		case <-c.done:
			return
		default:
		}

		if err := c.connect(ctx); err == nil {
			select {
			case c.reconnectC <- struct{}{}:
			default:
			}
			return
		}

		if backoff < 30*time.Second {
			backoff *= 2
		}
		logging.Error(ctx, errors.New(fmt.Sprintf("reconnect failed, waiting. backoff: %v", backoff)))
		time.Sleep(backoff)
	}
}

func (c *Client) Channel(ctx context.Context) (*amqp.Channel, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	ch, err := conn.Channel()
	if err != nil {
		err = fmt.Errorf("open channel: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}
	return ch, nil
}

func (c *Client) Close() error {
	close(c.done)
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if c.IsAlive() {
		return conn.Close()
	}
	return nil
}
func (c *Client) IsAlive() bool {
	return c.conn != nil && !c.conn.IsClosed()
}
