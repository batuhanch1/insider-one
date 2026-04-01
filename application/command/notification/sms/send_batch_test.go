package sms

import (
	"context"
	"errors"
	"testing"

	"insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendBatchSmsCommand_Execute_AllPriorities_PublishesThreeTimes(t *testing.T) {
	bp := &mockBatchPublisher{}
	bp.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	cmd := NewSendBatchCommand(bp)
	ctx := context.Background()

	req := SendBatchSmsRequest{
		Sms: []SendBatchSms{
			{PhoneNumber: "+15551111111", Sender: "S", Content: "C", Type: "T", Priority: "HIGH"},
			{PhoneNumber: "+15552222222", Sender: "S", Content: "C", Type: "T", Priority: "MEDIUM"},
			{PhoneNumber: "+15553333333", Sender: "S", Content: "C", Type: "T", Priority: "LOW"},
		},
	}

	err := cmd.Execute(ctx, req)
	assert.NoError(t, err)
	bp.AssertNumberOfCalls(t, "Publish", 3)
}

func TestSendBatchSmsCommand_Execute_HighPriorityError_ReturnsError(t *testing.T) {
	bp := &mockBatchPublisher{}
	bp.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.BatchPublisherOptions) bool {
		return opts.RoutingKey == rabbitmq.RoutingKey_High
	})).Return(errors.New("mq error")).Once()

	cmd := NewSendBatchCommand(bp)
	ctx := context.Background()
	req := SendBatchSmsRequest{
		Sms: []SendBatchSms{
			{PhoneNumber: "+15551111111", Sender: "S", Content: "C", Type: "T", Priority: "HIGH"},
		},
	}

	err := cmd.Execute(ctx, req)
	assert.Error(t, err)
	bp.AssertNumberOfCalls(t, "Publish", 1)
}
