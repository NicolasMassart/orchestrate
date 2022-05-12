package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/webhook"
)

type Client struct {
	c *http.Client
}

var _ webhook.Producer = &Client{}

func New(client *http.Client) *Client {
	return &Client{c: client}
}

func (c *Client) Send(ctx context.Context, specs *entities.EventStreamWebhookSpec, body interface{}) error {
	reqBody := new(bytes.Buffer)
	err := json.NewEncoder(reqBody).Encode(body)
	if err != nil {
		return err
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, specs.URL, reqBody)
	req.Header.Set("Content-Type", "application/json")
	for key, val := range specs.Headers {
		req.Header.Set(key, val)
	}

	_, err = c.c.Do(req)
	if err != nil {
		return err
	}

	return nil
}
