package sms

import (
	"context"
	"errors"
	"testing"

	sms_domain "insider-one/domain/notification/sms"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestScheduleSmsCommand_Execute_Success_PublishesForEachScheduled(t *testing.T) {
	repo := &mockSmsRepository{}
	pub := &mockPublisher{}

	smsList := sms_domain.SmsList{
		{ID: 1, Priority: "HIGH"},
		{ID: 2, Priority: "LOW"},
	}
	repo.On("GetScheduled", mock.Anything, mock.Anything).Return(smsList, nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
		return opts.Exchange == rabbitmq.Exchange_SmsCreated
	})).Return(nil)

	cmd := NewScheduleCommand(repo, pub)
	err := cmd.Execute(context.Background())
	assert.NoError(t, err)
	pub.AssertNumberOfCalls(t, "Publish", 2)
}

func TestScheduleSmsCommand_Execute_NoScheduled(t *testing.T) {
	repo := &mockSmsRepository{}
	pub := &mockPublisher{}

	repo.On("GetScheduled", mock.Anything, mock.Anything).Return(sms_domain.SmsList{}, nil)

	cmd := NewScheduleCommand(repo, pub)
	err := cmd.Execute(context.Background())
	assert.NoError(t, err)
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestScheduleSmsCommand_Execute_RepoError(t *testing.T) {
	repo := &mockSmsRepository{}
	pub := &mockPublisher{}

	repo.On("GetScheduled", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	cmd := NewScheduleCommand(repo, pub)
	err := cmd.Execute(context.Background())
	assert.Error(t, err)
}

func TestScheduleSmsCommand_Execute_PublisherError_ReturnsNil(t *testing.T) {
	repo := &mockSmsRepository{}
	pub := &mockPublisher{}

	repo.On("GetScheduled", mock.Anything, mock.Anything).Return(sms_domain.SmsList{{ID: 1, Priority: "HIGH"}}, nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("mq error"))

	cmd := NewScheduleCommand(repo, pub)
	err := cmd.Execute(context.Background())
	assert.NoError(t, err) // schedule command swallows publisher errors
}
