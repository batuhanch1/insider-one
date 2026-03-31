package email_provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"insider-one/infrastructure/adapters/client"
	"insider-one/infrastructure/config"
	"io"
	"net/http"
)

const (
	Path_Deliver = "/email/deliver"
)

type EmailProvider interface {
	Deliver(ctx context.Context, request *DeliverRequest) (*DeliverResponse, error)
}

type emailProvider struct {
	client client.HttpClient
	config config.EmailProviderConfig
}

func NewEmailProvider(client client.HttpClient, config config.EmailProviderConfig) EmailProvider {
	return &emailProvider{client, config}
}

func (s *emailProvider) Deliver(ctx context.Context, request *DeliverRequest) (*DeliverResponse, error) {
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("EmailProvider Deliver Marshal error %w", err)
	}

	url := fmt.Sprintf("%s%s", s.config.Host, Path_Deliver)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("EmailProvider Deliver NewRequest error %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.config.User, s.config.Password)

	var resp *http.Response
	if resp, err = s.client.Do(ctx, req); err != nil {
		return nil, fmt.Errorf("EmailProvider Deliver Do error %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("bad response: %d", resp.StatusCode)
	}

	var bodyByte []byte
	if bodyByte, err = io.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("EmailProvider Deliver ReadAll error %w", err)
	}

	var response DeliverResponse
	if err = json.Unmarshal(bodyByte, &response); err != nil {
		return nil, fmt.Errorf("EmailProvider Deliver Unmarshal error %w", err)
	}
	return &response, nil
}
