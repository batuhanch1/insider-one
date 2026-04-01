package sms

import (
	"context"
	"errors"
	"testing"
	"time"

	sms_domain "insider-one/domain/notification/sms"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSmsCommand_Execute_NotScheduled_SavesAndPublishes(t *testing.T) {
	repo := &mockSmsRepository{}
	pub := &mockPublisher{}

	repo.On("Save", mock.Anything, mock.Anything).Return(nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
		return opts.Exchange == rabbitmq.Exchange_SmsCreated
	})).Return(nil)

	cmd := NewCreateCommand(repo, pub)
	ctx := context.Background()
	event := sms_domain.CreateSmsEvent{
		PhoneNumber: "+15551234567",
		Sender:      "S",
		Type:        "T",
		Content:     "C",
		Priority:    "HIGH",
	}

	err := cmd.Execute(ctx, event)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Save", mock.Anything, mock.Anything)
	pub.AssertCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateSmsCommand_Execute_Scheduled_SavesAndSkipsPublish(t *testing.T) {
	repo := &mockSmsRepository{}
	pub := &mockPublisher{}

	repo.On("Save", mock.Anything, mock.Anything).Return(nil)

	cmd := NewCreateCommand(repo, pub)
	ctx := context.Background()
	event := sms_domain.CreateSmsEvent{
		ScheduledAt: time.Now().Unix() + 3600,
		PhoneNumber: "+15551234567",
		Sender:      "S",
		Type:        "T",
		Content:     "C",
	}

	err := cmd.Execute(ctx, event)
	assert.NoError(t, err)
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateSmsCommand_Execute_RepositoryError(t *testing.T) {
	repo := &mockSmsRepository{}
	pub := &mockPublisher{}

	repo.On("Save", mock.Anything, mock.Anything).Return(errors.New("db error"))

	cmd := NewCreateCommand(repo, pub)
	ctx := context.Background()
	event := sms_domain.CreateSmsEvent{PhoneNumber: "+1555", Sender: "S"}

	err := cmd.Execute(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	pub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateSmsCommand_Execute_PublisherError(t *testing.T) {
	repo := &mockSmsRepository{}
	pub := &mockPublisher{}

	repo.On("Save", mock.Anything, mock.Anything).Return(nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("mq error"))

	cmd := NewCreateCommand(repo, pub)
	ctx := context.Background()
	event := sms_domain.CreateSmsEvent{PhoneNumber: "+1555", Sender: "S"}

	err := cmd.Execute(ctx, event)
	assert.Error(t, err)
}
