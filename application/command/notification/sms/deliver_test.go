package sms

import (
	"context"
	"errors"
	"testing"

	sms_domain "insider-one/domain/notification/sms"
	sms_provider "insider-one/infrastructure/adapters/client/sms-provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeliverSmsCommand_Execute_Accepted_CallsRepoDeliver(t *testing.T) {
	repo := &mockSmsRepository{}
	prov := &mockSmsProvider{}

	var idempotencyKey uint64 = 99
	repo.On("Deliver", mock.Anything, "msg-sms-1", idempotencyKey).Return(nil)
	prov.On("Deliver", mock.Anything, mock.Anything).Return(
		&sms_provider.DeliverResponse{Status: "accepted", MessageID: "msg-sms-1"}, nil,
	)

	cmd := NewDeliverCommand(repo, prov, nil)
	ctx := context.Background()
	event := sms_domain.SmsCreatedEvent{
		PhoneNumber:    "+15551234567",
		Sender:         "S",
		Content:        "Hello",
		IdempotencyKey: idempotencyKey,
	}

	err := cmd.Execute(ctx, event)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Deliver", mock.Anything, "msg-sms-1", idempotencyKey)
}

func TestDeliverSmsCommand_Execute_NotAccepted_SkipsRepoDeliver(t *testing.T) {
	repo := &mockSmsRepository{}
	prov := &mockSmsProvider{}

	prov.On("Deliver", mock.Anything, mock.Anything).Return(
		&sms_provider.DeliverResponse{Status: "rejected"}, nil,
	)

	cmd := NewDeliverCommand(repo, prov, nil)
	ctx := context.Background()
	event := sms_domain.SmsCreatedEvent{PhoneNumber: "+1555", Sender: "S"}

	err := cmd.Execute(ctx, event)
	assert.NoError(t, err)
	repo.AssertNotCalled(t, "Deliver", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeliverSmsCommand_Execute_ProviderError(t *testing.T) {
	repo := &mockSmsRepository{}
	prov := &mockSmsProvider{}

	prov.On("Deliver", mock.Anything, mock.Anything).Return(nil, errors.New("provider down"))

	cmd := NewDeliverCommand(repo, prov, nil)
	ctx := context.Background()
	event := sms_domain.SmsCreatedEvent{PhoneNumber: "+1555", Sender: "S"}

	err := cmd.Execute(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider down")
}

func TestDeliverSmsCommand_Execute_RepoDeliverError(t *testing.T) {
	repo := &mockSmsRepository{}
	prov := &mockSmsProvider{}

	prov.On("Deliver", mock.Anything, mock.Anything).Return(
		&sms_provider.DeliverResponse{Status: "accepted", MessageID: "msg-2"}, nil,
	)
	repo.On("Deliver", mock.Anything, "msg-2", mock.Anything).Return(errors.New("db error"))

	cmd := NewDeliverCommand(repo, prov, nil)
	ctx := context.Background()
	event := sms_domain.SmsCreatedEvent{PhoneNumber: "+1555", Sender: "S"}

	err := cmd.Execute(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
