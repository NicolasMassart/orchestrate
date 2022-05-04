package sarama

import (
	"encoding/json"
	"time"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/infra/kafka"
	"github.com/hashicorp/go-multierror"
	healthz "github.com/heptiolabs/healthcheck"
)

type Client struct {
	syncProducer sarama.SyncProducer
	addrs        []string
}

var _ kafka.Producer = &Client{}

func NewProducer(cfg *Config) (*Client, error) {
	saramaCfg, err := cfg.ToSaramaConfig()
	if err != nil {
		return nil, err
	}

	p, err := sarama.NewSyncProducer(cfg.URLs, saramaCfg)
	if err != nil {
		return nil, err
	}

	return &Client{syncProducer: p, addrs: cfg.URLs}, nil
}

func (p *Client) Send(body interface{}, topic, partitionKey string, headers map[string]interface{}) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
	}

	if partitionKey != "" {
		msg.Key = sarama.StringEncoder(partitionKey)
	}

	bValue, err := json.Marshal(body)
	if err != nil {
		return errors.EncodingError("failed to marshall message body")
	}

	msg.Value = sarama.ByteEncoder(bValue)

	if headers != nil {
		msg.Headers = []sarama.RecordHeader{}

		for headerKey, headerValue := range headers {
			bHeaderValue, _ := json.Marshal(headerValue)
			msg.Headers = append(msg.Headers, sarama.RecordHeader{
				Key:   []byte(headerKey),
				Value: bHeaderValue,
			})
		}
	}

	_, _, err = p.syncProducer.SendMessage(msg)
	if err != nil {
		return errors.KafkaConnectionError("could not send request message. %s", err.Error())
	}

	return nil
}

func (p *Client) Close() error {
	return p.syncProducer.Close()
}

func (p *Client) Checker() error {
	gr := &multierror.Group{}
	for _, host := range p.addrs {
		gr.Go(healthz.TCPDialCheck(host, time.Second*3))
	}

	return gr.Wait().ErrorOrNil()
}
