package sarama

import (
	encoding "encoding/json"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/kafka"
)

type Producer struct {
	syncProducer sarama.SyncProducer
	client       kafka.Client
}

var _ kafka.Producer = &Producer{}

func NewProducer(cfg *Config) (*Producer, error) {
	saramaCfg, err := cfg.ToKafkaConfig()
	if err != nil {
		return nil, err
	}

	client, err := NewClient(saramaCfg, cfg.URLs)
	if err != nil {
		return nil, err
	}

	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, errors.KafkaConnectionError(err.Error())
	}

	return &Producer{syncProducer: p, client: client}, nil
}

func (p *Producer) SendJobMessage(topic string, job *entities.Job, userInfo *multitenancy.UserInfo) error { // TODO(dario): remove this function from the kakfa folder to an internal notifier
	msg := &sarama.ProducerMessage{
		Topic: topic,
	}

	partitionKey := job.PartitionKey()
	if partitionKey != "" {
		msg.Key = sarama.StringEncoder(partitionKey)
	}

	// Marshal protobuffer into byte
	b, err := encoding.Marshal(job)
	if err != nil {
		return errors.EncodingError("failed to marshall envelope")
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

func (p *Producer) Client() kafka.Client {
	return p.client
}
