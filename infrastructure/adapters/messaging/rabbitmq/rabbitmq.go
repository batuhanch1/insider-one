package rabbitmq

import (
	"fmt"
	"insider-one/infrastructure/config"
	"log/slog"
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

func New(cfg *config.Config) (*Client, error) {
	c := &Client{
		cfg:        *cfg,
		done:       make(chan struct{}),
		reconnectC: make(chan struct{}, 1),
	}
	if err := c.connect(); err != nil {
		return nil, err
	}
	go c.watchConnection()
	return c, nil
}

func (c *Client) connect() error {
	url := fmt.Sprintf("AMQP://%s:%s@%s:%d/", c.cfg.Rabbit.User, c.cfg.Rabbit.Password, c.cfg.Rabbit.Host, c.cfg.Rabbit.Port)
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("AMQP dial: %w", err)
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	return nil
}

func (c *Client) watchConnection() {
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
			slog.Warn("RabbitMQ connection lost, reconnecting",
				"error", amqpErr,
			)
			c.reconnectWithBackoff()

		case <-c.done:
			return
		}
	}
}

func (c *Client) reconnectWithBackoff() {
	backoff := 1 * time.Second
	for attempt := 1; ; attempt++ {
		select {
		case <-c.done:
			return
		default:
		}

		slog.Info("attempting reconnect", "attempt", attempt)
		if err := c.connect(); err == nil {
			slog.Info("RabbitMQ reconnected", "attempt", attempt)
			select {
			case c.reconnectC <- struct{}{}:
			default:
			}
			return
		}

		if backoff < 30*time.Second {
			backoff *= 2
		}
		slog.Warn("reconnect failed, waiting", "backoff", backoff)
		time.Sleep(backoff)
	}
}

func (c *Client) Channel() (*amqp.Channel, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}
	return ch, nil
}

func (c *Client) Close() error {
	close(c.done)
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn != nil && !conn.IsClosed() {
		return conn.Close()
	}
	return nil
}
func (c *Client) IsAlive() bool {
	return c.conn != nil && !c.conn.IsClosed()
}
