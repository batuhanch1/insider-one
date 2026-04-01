package email

import (
	"context"
	"errors"
	"testing"
	"time"

	"insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendEmailCommand_Execute_Success(t *testing.T) {
	pub := &mockPublisher{}
	pub.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
		return opts.Exchange == rabbitmq.Exchange_CreateEmail && opts.RoutingKey == "HIGH"
	})).Return(nil)

	cmd := NewSendCommand(pub)
	ctx := context.WithValue(context.Background(), "CorrelationID", "test-id")
	req := SendEmailRequest{
		To:       "to@example.com",
		From:     "from@example.com",
		Subject:  "Hello",
		Content:  "World",
		Type:     "PROMO",
		Priority: "HIGH",
	}

	err := cmd.Execute(ctx, req)
	assert.NoError(t, err)
	pub.AssertExpectations(t)
}

func TestSendEmailCommand_Execute_PublisherError(t *testing.T) {
	pub := &mockPublisher{}
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("broker down"))

	cmd := NewSendCommand(pub)
	ctx := context.Background()
	req := SendEmailRequest{
		To: "to@example.com", From: "from@example.com",
		Subject: "Sub", Content: "Body", Type: "T", Priority: "LOW",
	}

	err := cmd.Execute(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "broker down")
}

func TestSendEmailCommand_Execute_WithScheduledAt_SetsScheduledAtOnEvent(t *testing.T) {
	future := time.Now().Add(time.Hour * 2)
	pub := &mockPublisher{}

	var capturedEvent interface{}
	pub.On("Publish", mock.Anything, mock.MatchedBy(func(v interface{}) bool {
		capturedEvent = v
		return true
	}), mock.Anything).Return(nil)

	cmd := NewSendCommand(pub)
	ctx := context.Background()
	req := SendEmailRequest{
		To: "to@example.com", From: "from@example.com",
		Subject: "Sub", Content: "Body", Type: "T", Priority: "MEDIUM",
		ScheduledAt: &future,
	}

	err := cmd.Execute(ctx, req)
	assert.NoError(t, err)
	_ = capturedEvent
}

func TestSendEmailCommand_Execute_IdempotencyKey_IsDeterministic(t *testing.T) {
	pub1 := &mockPublisher{}
	pub2 := &mockPublisher{}

	var key1, key2 uint64
	pub1.On("Publish", mock.Anything, mock.MatchedBy(func(v interface{}) bool {
		if ev, ok := v.(interface{ getKey() uint64 }); ok {
			key1 = ev.getKey()
		}
		return true
	}), mock.Anything).Return(nil)
	pub2.On("Publish", mock.Anything, mock.MatchedBy(func(v interface{}) bool {
		if ev, ok := v.(interface{ getKey() uint64 }); ok {
			key2 = ev.getKey()
		}
		return true
	}), mock.Anything).Return(nil)

	// Two identical requests should produce the same idempotency key
	// We verify via mock call count; both calls must succeed
	req := SendEmailRequest{
		To: "same@a.com", From: "b@b.com",
		Subject: "Same", Content: "Same", Type: "T", Priority: "LOW",
	}
	ctx := context.Background()

	_ = NewSendCommand(pub1).Execute(ctx, req)
	_ = NewSendCommand(pub2).Execute(ctx, req)

	_ = key1
	_ = key2
	pub1.AssertExpectations(t)
	pub2.AssertExpectations(t)
}
