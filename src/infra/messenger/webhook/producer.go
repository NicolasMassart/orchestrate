package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/infra/messenger"
	"github.com/consensys/orchestrate/src/infra/messenger/types"
	types2 "github.com/consensys/orchestrate/src/notifier/service/types"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
)

const notSupportedErr = "not supported"

type Producer struct {
	client *http.Client
}

var _ messenger.Producer = &Producer{}

// @TODO Replace by config

func NewProducer(client *http.Client) *Producer {
	return &Producer{
		client: client,
	}
}

func (c *Producer) SendNotificationResponse(ctx context.Context, notif *entities.Notification, eventStream *entities.EventStream) error {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(types.NewNotificationResponse(notif))
	if err != nil {
		return errors.EncodingError(err.Error())
	}

	webhook := eventStream.Webhook

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, body)
	req.Header.Set("Content-Type", "application/json")
	for key, val := range webhook.Headers {
		req.Header.Set(key, val)
	}

	_, err = c.client.Do(req)
	if err != nil {
		return errors.HTTPConnectionError(err.Error())
	}

	return nil
}

func (c *Producer) SendJobMessage(_ string, _ *entities.Job, _ string, _ *multitenancy.UserInfo) error {
	return errors.FeatureNotSupportedError(notSupportedErr)
}

func (c *Producer) SendNotificationMessage(_ string, _ *types2.NotificationMessage, _ string, _ *multitenancy.UserInfo) error {
	return errors.FeatureNotSupportedError(notSupportedErr)
}

func (c *Producer) Checker() error {
	return errors.FeatureNotSupportedError(notSupportedErr)
}
