package sms_provider

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

type SmsProvider interface {
	Deliver(ctx context.Context, request *DeliverRequest) (*DeliverResponse, error)
}

type smsProvider struct {
	client client.HttpClient
	config config.SmsProviderConfig
}

func NewSmsProvider(client client.HttpClient, config config.SmsProviderConfig) SmsProvider {
	return &smsProvider{client, config}
}

func (s *smsProvider) Deliver(ctx context.Context, request *DeliverRequest) (*DeliverResponse, error) {
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("smsProvider Deliver Marshal error %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	url := fmt.Sprintf("%s%s", s.config.Host, Path_Deliver)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		err = fmt.Errorf("smsProvider Deliver NewRequest error %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.config.User, s.config.Password)

	var resp *http.Response
	if resp, err = s.client.Do(ctx, req); err != nil {
		err = fmt.Errorf("smsProvider Deliver Do error %w", err)
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
		err = fmt.Errorf("smsProvider Deliver ReadAll error %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	var response DeliverResponse
	if err = json.Unmarshal(bodyByte, &response); err != nil {
		err = fmt.Errorf("smsProvider Deliver Unmarshal error %w", err)
		logging.Error(ctx, err)
		return nil, err
	}
	return &response, nil
}
