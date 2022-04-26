package kafka

import (
	encoding "encoding/json"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/kafka"
	saramainfra "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/messenger"
)

type Client struct {
	syncProducer sarama.SyncProducer
	client       kafka.Client
}

var _ messenger.Producer = &Client{}

func NewProducer(cfg *saramainfra.Config) (*Client, error) {
	client, err := saramainfra.NewClient(cfg, cfg.URLs)
	if err != nil {
		return nil, err
	}

	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, errors.KafkaConnectionError(err.Error())
	}

	return &Client{syncProducer: p, client: client}, nil
}

func (p *Client) SendJobMessage(topic string, job *entities.Job, userInfo *multitenancy.UserInfo) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
	}

	partitionKey := job.PartitionKey()
	if partitionKey != "" {
		msg.Key = sarama.StringEncoder(partitionKey)
	}

	b, err := encoding.Marshal(job)
	if err != nil {
		return errors.EncodingError("failed to marshall job message")
	}
	msg.Value = sarama.ByteEncoder(b)

	if userInfo.AuthMode == multitenancy.AuthMethodJWT {
		userInfoB, _ := encoding.Marshal(userInfo)
		msg.Headers = []sarama.RecordHeader{
			{
				Key:   []byte(utils.UserInfoHeader),
				Value: userInfoB,
			},
		}
	}

	_, _, err = p.syncProducer.SendMessage(msg)
	if err != nil {
		errMsg := "could not produce kafka message"
		return errors.KafkaConnectionError(errMsg)
	}

	return nil
}

func (p *Client) Checker() error {
	return p.client.Checker()
}
