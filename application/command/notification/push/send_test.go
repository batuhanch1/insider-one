package push

import (
	"context"
	"errors"
	"testing"

	"insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendPushCommand_Execute_Success(t *testing.T) {
	pub := &mockPublisher{}
	pub.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
		return opts.Exchange == rabbitmq.Exchange_CreatePush && opts.RoutingKey == "HIGH"
	})).Return(nil)

	cmd := NewSendCommand(pub)
	ctx := context.Background()
	req := SendPushRequest{
		PhoneNumber: "+15551234567",
		Sender:      "PUSH",
		Content:     "Notification",
		Type:        "ALERT",
		Priority:    "HIGH",
	}

	err := cmd.Execute(ctx, req)
	assert.NoError(t, err)
	pub.AssertExpectations(t)
}

func TestSendPushCommand_Execute_PublisherError(t *testing.T) {
	pub := &mockPublisher{}
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("broker down"))

	cmd := NewSendCommand(pub)
	ctx := context.Background()
	req := SendPushRequest{PhoneNumber: "+15551234567", Sender: "P", Content: "C", Type: "T", Priority: "LOW"}

	err := cmd.Execute(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "broker down")
}
