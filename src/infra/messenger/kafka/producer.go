package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/consensys/orchestrate/src/infra/messenger/types"
	notifier "github.com/consensys/orchestrate/src/notifier/service/types"
	"github.com/hashicorp/go-multierror"
	healthz "github.com/heptiolabs/healthcheck"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/messenger"
)

type Client struct {
	syncProducer sarama.SyncProducer
	addrs        []string
}

var _ messenger.Producer = &Client{}

func NewProducer(cfg *Config) (*Client, error) {
	saramaCfg, err := cfg.ToKafkaConfig()
	if err != nil {
		return nil, err
	}

	p, err := sarama.NewSyncProducer(cfg.URLs, saramaCfg)
	if err != nil {
		return nil, err
	}

	return &Client{syncProducer: p, addrs: cfg.URLs}, nil
}

func (p *Client) SendJobMessage(topic string, job *entities.Job, partitionKey string, userInfo *multitenancy.UserInfo) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
	}

	if partitionKey != "" {
		msg.Key = sarama.StringEncoder(partitionKey)
	}

	b, err := json.Marshal(job)
	if err != nil {
		return errors.EncodingError("failed to marshall job message")
	}
	msg.Value = sarama.ByteEncoder(b)

	if userInfo.AuthMode == multitenancy.AuthMethodJWT {
		userInfoB, _ := json.Marshal(userInfo)
		msg.Headers = []sarama.RecordHeader{
			{
				Key:   []byte(utils.UserInfoHeader),
				Value: userInfoB,
			},
		}
	}

	_, _, err = p.syncProducer.SendMessage(msg)
	if err != nil {
		return errors.KafkaConnectionError("could not send job message")
	}

	return nil
}

func (p *Client) SendNotificationMessage(topic string, notif *notifier.NotificationMessage, partitionKey string, userInfo *multitenancy.UserInfo) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
	}

	if partitionKey != "" {
		msg.Key = sarama.StringEncoder(partitionKey)
	}

	b, err := json.Marshal(notif)
	if err != nil {
		return err
	}
	msg.Value = sarama.ByteEncoder(b)

	if userInfo.AuthMode == multitenancy.AuthMethodJWT {
		userInfoB, _ := json.Marshal(userInfo)
		msg.Headers = []sarama.RecordHeader{
			{
				Key:   []byte(utils.UserInfoHeader),
				Value: userInfoB,
			},
		}
	}

	_, _, err = p.syncProducer.SendMessage(msg)
	if err != nil {
		return errors.KafkaConnectionError(err.Error())
	}

	return nil
}

func (p *Client) SendNotificationResponse(_ context.Context, notif *entities.Notification, eventStream *entities.EventStream) error {
	body, err := json.Marshal(types.NewNotificationResponse(notif))
	if err != nil {
		return errors.EncodingError(err.Error())
	}

	_, _, err = p.syncProducer.SendMessage(&sarama.ProducerMessage{Topic: eventStream.Kafka.Topic, Value: sarama.ByteEncoder(body)})
	if err != nil {
		return errors.KafkaConnectionError("could not send notification response")
	}

	return nil
}

func (p *Client) Checker() error {
	gr := &multierror.Group{}
	for _, host := range p.addrs {
		gr.Go(healthz.TCPDialCheck(host, time.Second*3))
	}

	return gr.Wait().ErrorOrNil()
}
