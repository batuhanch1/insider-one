package client

import (
	"context"
	"insider-one/infrastructure/logging"
	"net/http"
	"time"
)

type HttpClient interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}
type httpClient struct {
	client *http.Client
}

func NewClient() HttpClient {
	return &httpClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
func (h *httpClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	logging.InternalLogStart(ctx, req)
	response, err := h.Do(ctx, req)
	logging.InternalLogFinish(ctx, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
