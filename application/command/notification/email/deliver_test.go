package email

import (
	"context"
	"errors"
	"testing"

	email_domain "insider-one/domain/notification/email"
	email_provider "insider-one/infrastructure/adapters/client/email-provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeliverEmailCommand_Execute_Accepted_CallsRepoDeliver(t *testing.T) {
	repo := &mockEmailRepository{}
	prov := &mockEmailProvider{}

	var idempotencyKey uint64 = 12345
	repo.On("Deliver", mock.Anything, "msg-id-1", idempotencyKey).Return(nil)
	prov.On("Deliver", mock.Anything, mock.Anything).Return(
		&email_provider.DeliverResponse{Status: "accepted", MessageID: "msg-id-1"}, nil,
	)

	// Publisher field is unused in Execute; pass nil
	cmd := NewDeliverCommand(repo, prov, nil)
	ctx := context.Background()
	event := email_domain.EmailCreatedEvent{
		To:             "to@a.com",
		From:           "from@a.com",
		Subject:        "Sub",
		Content:        "Body",
		IdempotencyKey: idempotencyKey,
	}

	err := cmd.Execute(ctx, event)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Deliver", mock.Anything, "msg-id-1", idempotencyKey)
}

func TestDeliverEmailCommand_Execute_NotAccepted_SkipsRepoDeliver(t *testing.T) {
	repo := &mockEmailRepository{}
	prov := &mockEmailProvider{}

	prov.On("Deliver", mock.Anything, mock.Anything).Return(
		&email_provider.DeliverResponse{Status: "rejected"}, nil,
	)

	cmd := NewDeliverCommand(repo, prov, nil)
	ctx := context.Background()
	event := email_domain.EmailCreatedEvent{To: "t@a.com", From: "f@a.com"}

	err := cmd.Execute(ctx, event)
	assert.NoError(t, err)
	repo.AssertNotCalled(t, "Deliver", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeliverEmailCommand_Execute_ProviderError_ReturnsError(t *testing.T) {
	repo := &mockEmailRepository{}
	prov := &mockEmailProvider{}

	prov.On("Deliver", mock.Anything, mock.Anything).Return(nil, errors.New("provider down"))

	cmd := NewDeliverCommand(repo, prov, nil)
	ctx := context.Background()
	event := email_domain.EmailCreatedEvent{To: "t@a.com", From: "f@a.com"}

	err := cmd.Execute(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider down")
	repo.AssertNotCalled(t, "Deliver", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeliverEmailCommand_Execute_RepoDeliverError_ReturnsError(t *testing.T) {
	repo := &mockEmailRepository{}
	prov := &mockEmailProvider{}

	prov.On("Deliver", mock.Anything, mock.Anything).Return(
		&email_provider.DeliverResponse{Status: "accepted", MessageID: "msg-2"}, nil,
	)
	repo.On("Deliver", mock.Anything, "msg-2", mock.Anything).Return(errors.New("db error"))

	cmd := NewDeliverCommand(repo, prov, nil)
	ctx := context.Background()
	event := email_domain.EmailCreatedEvent{To: "t@a.com", From: "f@a.com"}

	err := cmd.Execute(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
