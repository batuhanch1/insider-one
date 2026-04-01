package email_provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"insider-one/infrastructure/adapters/client"
	"insider-one/infrastructure/config"
	"insider-one/infrastructure/logging"
	"io"
	"net/http"
)

const (
	Path_Deliver = ""
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
		err = fmt.Errorf("EmailProvider Deliver Marshal error %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	url := fmt.Sprintf("%s%s", s.config.Host, Path_Deliver)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		err = fmt.Errorf("EmailProvider Deliver NewRequest error %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.config.User, s.config.Password)

	var resp *http.Response
	if resp, err = s.client.Do(ctx, req); err != nil {
		err = fmt.Errorf("EmailProvider Deliver Do error %w", err)
		logging.Error(ctx, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		err = fmt.Errorf("bad response: %d", resp.StatusCode)
		logging.Error(ctx, err)
		return nil, err
	}

	var bodyByte []byte
	if bodyByte, err = io.ReadAll(resp.Body); err != nil {
		err = fmt.Errorf("EmailProvider Deliver ReadAll error %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	var response DeliverResponse
	if err = json.Unmarshal(bodyByte, &response); err != nil {
		err = fmt.Errorf("EmailProvider Deliver Unmarshal error %w", err)
		logging.Error(ctx, err)
		return nil, err
	}
	return &response, nil
}
