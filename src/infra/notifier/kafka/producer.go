package kafka

import (
	"context"
	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
	saramainfra "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/notifier"
	notification "github.com/consensys/orchestrate/src/infra/notifier/types"
)

type Producer struct {
	syncProducer sarama.SyncProducer
}

var _ notifier.Producer = &Producer{}

func NewProducer(cfg *saramainfra.Config) (*Producer, error) {
	saramaCfg, err := cfg.ToKafkaConfig()
	if err != nil {
		return nil, err
	}

	p, err := sarama.NewSyncProducer(cfg.URLs, saramaCfg)
	if err != nil {
		return nil, err
	}

	return &Producer{syncProducer: p}, nil
}

func (c *Producer) SendTxResponse(ctx context.Context, eventStream *entities.EventStream, job *entities.Job, errStr string) error {
	notif := notification.NewTxNotification(job, errStr)

	body, err := json.Marshal(notif)
	if err != nil {
		return errors.EncodingError(err.Error())
	}

	msg := &sarama.ProducerMessage{
		Topic: eventStream.Kafka().Topic,
		Key:   sarama.StringEncoder(job.ChainUUID), // todo(Dario) decide on a coherent partition key
		Value: sarama.ByteEncoder(body),
	}

	_, _, err = c.syncProducer.SendMessage(msg)
	if err != nil {
		return errors.KafkaConnectionError("could not produce kafka message")
	}

	return nil
}
