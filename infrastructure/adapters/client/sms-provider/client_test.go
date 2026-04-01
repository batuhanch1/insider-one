package sms_provider

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"insider-one/infrastructure/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockHttpClient struct {
	mock.Mock
}

func (m *mockHttpClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func newTestProvider(c *mockHttpClient) SmsProvider {
	return NewSmsProvider(c, config.SmsProviderConfig{
		Host:     "http://test.example.com",
		User:     "user",
		Password: "pass",
	})
}

func TestSmsProvider_Deliver_Success(t *testing.T) {
	mockClient := &mockHttpClient{}
	body := `{"message_id":"msg-1","status":"accepted"}`
	resp := &http.Response{
		StatusCode: http.StatusAccepted,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
	mockClient.On("Do", mock.Anything, mock.Anything).Return(resp, nil)

	provider := newTestProvider(mockClient)
	result, err := provider.Deliver(context.Background(), &DeliverRequest{
		PhoneNumber: "+905551234567", Sender: "TEST",
	})

	assert.NoError(t, err)
	assert.Equal(t, "msg-1", result.MessageID)
	assert.Equal(t, "accepted", result.Status)
	mockClient.AssertExpectations(t)
}

func TestSmsProvider_Deliver_HttpClientError_ReturnsError(t *testing.T) {
	mockClient := &mockHttpClient{}
	mockClient.On("Do", mock.Anything, mock.Anything).Return(nil, errors.New("connection refused"))

	provider := newTestProvider(mockClient)
	result, err := provider.Deliver(context.Background(), &DeliverRequest{})

	assert.Error(t, err)
	assert.Nil(t, result)
	mockClient.AssertExpectations(t)
}

func TestSmsProvider_Deliver_BadStatusCode_ReturnsError(t *testing.T) {
	mockClient := &mockHttpClient{}
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}
	mockClient.On("Do", mock.Anything, mock.Anything).Return(resp, nil)

	provider := newTestProvider(mockClient)
	result, err := provider.Deliver(context.Background(), &DeliverRequest{})

	assert.Error(t, err)
	assert.Nil(t, result)
	mockClient.AssertExpectations(t)
}

func TestSmsProvider_Deliver_UnmarshalError_ReturnsError(t *testing.T) {
	mockClient := &mockHttpClient{}
	resp := &http.Response{
		StatusCode: http.StatusAccepted,
		Body:       io.NopCloser(bytes.NewBufferString("not-json")),
	}
	mockClient.On("Do", mock.Anything, mock.Anything).Return(resp, nil)

	provider := newTestProvider(mockClient)
	result, err := provider.Deliver(context.Background(), &DeliverRequest{})

	assert.Error(t, err)
	assert.Nil(t, result)
	mockClient.AssertExpectations(t)
}
