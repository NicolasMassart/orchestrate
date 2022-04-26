package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/notifier"
	notification "github.com/consensys/orchestrate/src/infra/notifier/types"
)

type Producer struct {
	client *http.Client
}

var _ notifier.Producer = &Producer{}

// @TODO Replace by config
func NewProducer(client *http.Client) *Producer {
	return &Producer{
		client: client,
	}
}

func (c *Producer) SendTxResponse(ctx context.Context, eventStream *entities.EventStream, job *entities.Job, errStr string) error {
	notif := notification.NewTxNotification(job, errStr)

	webhookSpec := eventStream.WebHook()
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(notif)
	if err != nil {
		return errors.EncodingError(err.Error())
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, webhookSpec.URL, body)
	req.Header.Set("Content-Type", "application/json")
	for key, val := range webhookSpec.Headers {
		req.Header.Set(key, val)
	}

	_, err = c.client.Do(req)
	if err != nil {
		return errors.HTTPConnectionError(err.Error())
	}

	return nil
}
