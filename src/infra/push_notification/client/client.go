package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
	saramainfra "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	pushnotification "github.com/consensys/orchestrate/src/infra/push_notification"
)

type Client struct {
	syncProducer sarama.SyncProducer
}

var _ pushnotification.Notifier = &Client{}

func New(cfg *saramainfra.Config) (*Client, error) {
	saramaCfg, err := cfg.ToKafkaConfig()
	if err != nil {
		return nil, err
	}

	p, err := sarama.NewSyncProducer(cfg.URLs, saramaCfg)
	if err != nil {
		return nil, err
	}

	return &Client{syncProducer: p}, nil
}

func (c *Client) SendTxResponse(ctx context.Context, eventStream *entities.EventStream, job *entities.Job, errStr string) error {
	switch eventStream.Channel {
	case entities.EventStreamChannelKafka:
		return c.sendKafka(ctx, eventStream.Kafka(), job, errStr)
	case entities.EventStreamChannelWebhook:
		return c.sendWebhook(ctx, eventStream.WebHook(), job, errStr)
	default:
		return errors.DataCorruptedError("invalid event stream channel")
	}
}

func (c *Client) sendWebhook(ctx context.Context, webhook *entities.Webhook, job *entities.Job, errStr string) error {
	notif := NewTxNotification(job, errStr)

	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(notif)
	if err != nil {
		return errors.EncodingError(err.Error())
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, body)
	req.Header.Set("Content-Type", "application/json")
	for key, val := range webhook.Headers {
		req.Header.Set(key, val)
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return errors.HTTPConnectionError(err.Error())
	}

	return nil
}

func (c *Client) sendKafka(_ context.Context, kafka *entities.Kafka, job *entities.Job, errStr string) error {
	notif := NewTxNotification(job, errStr)

	body, err := json.Marshal(notif)
	if err != nil {
		return errors.EncodingError(err.Error())
	}

	msg := &sarama.ProducerMessage{
		Topic: kafka.Topic,
		Key:   sarama.StringEncoder(job.ChainUUID), // todo(Dario) decide on a coherent partition key
		Value: sarama.ByteEncoder(body),
	}

	_, _, err = c.syncProducer.SendMessage(msg)
	if err != nil {
		return errors.KafkaConnectionError("could not produce kafka message")
	}

	return nil
}
