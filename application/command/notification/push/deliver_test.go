package push

import (
	"context"
	"errors"
	"testing"

	push_domain "insider-one/domain/notification/push"
	push_provider "insider-one/infrastructure/adapters/client/push-provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeliverPushCommand_Execute_Accepted_CallsRepoDeliver(t *testing.T) {
	repo := &mockPushRepository{}
	prov := &mockPushProvider{}

	var idempotencyKey uint64 = 77
	repo.On("Deliver", mock.Anything, "msg-push-1", idempotencyKey).Return(nil)
	prov.On("Deliver", mock.Anything, mock.Anything).Return(
		&push_provider.DeliverResponse{Status: "accepted", MessageID: "msg-push-1"}, nil,
	)

	cmd := NewDeliverCommand(repo, prov, nil)
	ctx := context.Background()
	event := push_domain.PushCreatedEvent{
		Sender:         "PUSH",
		PhoneNumber:    "+15551234567",
		Content:        "Hello",
		IdempotencyKey: idempotencyKey,
	}

	err := cmd.Execute(ctx, event)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Deliver", mock.Anything, "msg-push-1", idempotencyKey)
}

func TestDeliverPushCommand_Execute_NotAccepted_SkipsRepoDeliver(t *testing.T) {
	repo := &mockPushRepository{}
	prov := &mockPushProvider{}

	prov.On("Deliver", mock.Anything, mock.Anything).Return(
		&push_provider.DeliverResponse{Status: "rejected"}, nil,
	)

	cmd := NewDeliverCommand(repo, prov, nil)
	ctx := context.Background()
	event := push_domain.PushCreatedEvent{Sender: "P", PhoneNumber: "+1555"}

	err := cmd.Execute(ctx, event)
	assert.NoError(t, err)
	repo.AssertNotCalled(t, "Deliver", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeliverPushCommand_Execute_ProviderError(t *testing.T) {
	repo := &mockPushRepository{}
	prov := &mockPushProvider{}

	prov.On("Deliver", mock.Anything, mock.Anything).Return(nil, errors.New("provider down"))

	cmd := NewDeliverCommand(repo, prov, nil)
	ctx := context.Background()
	event := push_domain.PushCreatedEvent{Sender: "P", PhoneNumber: "+1555"}

	err := cmd.Execute(ctx, event)
	assert.Error(t, err)
}

func TestDeliverPushCommand_Execute_RepoDeliverError(t *testing.T) {
	repo := &mockPushRepository{}
	prov := &mockPushProvider{}

	prov.On("Deliver", mock.Anything, mock.Anything).Return(
		&push_provider.DeliverResponse{Status: "accepted", MessageID: "msg-3"}, nil,
	)
	repo.On("Deliver", mock.Anything, "msg-3", mock.Anything).Return(errors.New("db error"))

	cmd := NewDeliverCommand(repo, prov, nil)
	ctx := context.Background()
	event := push_domain.PushCreatedEvent{Sender: "P", PhoneNumber: "+1555"}

	err := cmd.Execute(ctx, event)
	assert.Error(t, err)
}
