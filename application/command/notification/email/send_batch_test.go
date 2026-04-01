package email

import (
	"context"
	"errors"
	"testing"

	"insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendBatchEmailCommand_Execute_AllPriorities_PublishesThreeTimes(t *testing.T) {
	bp := &mockBatchPublisher{}
	bp.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	cmd := NewSendBatchCommand(bp)
	ctx := context.Background()

	req := SendBatchEmailRequest{
		Emails: []SendBatchEmail{
			{To: "a@a.com", From: "b@b.com", Subject: "S", Content: "C", Type: "T", Priority: "HIGH"},
			{To: "c@c.com", From: "d@d.com", Subject: "S", Content: "C", Type: "T", Priority: "MEDIUM"},
			{To: "e@e.com", From: "f@f.com", Subject: "S", Content: "C", Type: "T", Priority: "LOW"},
		},
	}

	err := cmd.Execute(ctx, req)
	assert.NoError(t, err)
	// publishBatch always called 3 times (high/medium/low), even for empty lists
	bp.AssertNumberOfCalls(t, "Publish", 3)
}

func TestSendBatchEmailCommand_Execute_OnlyHighPriority_StillCallsPublishThreeTimes(t *testing.T) {
	bp := &mockBatchPublisher{}
	bp.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	cmd := NewSendBatchCommand(bp)
	ctx := context.Background()

	req := SendBatchEmailRequest{
		Emails: []SendBatchEmail{
			{To: "a@a.com", From: "b@b.com", Subject: "S", Content: "C", Type: "T", Priority: "HIGH"},
		},
	}

	err := cmd.Execute(ctx, req)
	assert.NoError(t, err)
	// All 3 publishBatch calls happen regardless of list content
	bp.AssertNumberOfCalls(t, "Publish", 3)
}

func TestSendBatchEmailCommand_Execute_HighPriorityPublisherError_ReturnsError(t *testing.T) {
	bp := &mockBatchPublisher{}
	bp.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.BatchPublisherOptions) bool {
		return opts.RoutingKey == rabbitmq.RoutingKey_High
	})).Return(errors.New("mq error")).Once()

	cmd := NewSendBatchCommand(bp)
	ctx := context.Background()

	req := SendBatchEmailRequest{
		Emails: []SendBatchEmail{
			{To: "a@a.com", From: "b@b.com", Subject: "S", Content: "C", Type: "T", Priority: "HIGH"},
		},
	}

	err := cmd.Execute(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mq error")
	bp.AssertNumberOfCalls(t, "Publish", 1)
}

func TestSendBatchEmailCommand_Execute_MediumPriorityError_ReturnsError(t *testing.T) {
	bp := &mockBatchPublisher{}
	// High succeeds, medium fails
	bp.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.BatchPublisherOptions) bool {
		return opts.RoutingKey == rabbitmq.RoutingKey_High
	})).Return(nil).Once()
	bp.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.BatchPublisherOptions) bool {
		return opts.RoutingKey == rabbitmq.RoutingKey_Medium
	})).Return(errors.New("mq medium error")).Once()

	cmd := NewSendBatchCommand(bp)
	ctx := context.Background()

	req := SendBatchEmailRequest{
		Emails: []SendBatchEmail{
			{To: "a@a.com", From: "b@b.com", Subject: "S", Content: "C", Type: "T", Priority: "MEDIUM"},
		},
	}

	err := cmd.Execute(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mq medium error")
}
