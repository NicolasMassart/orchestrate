package jobs

import (
	"github.com/Shopify/sarama"
	saramaencoding "github.com/consensys/orchestrate/pkg/encoding/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/src/entities"
)

func sendEnvelope(kafkaProducer sarama.SyncProducer, topic string, job *entities.Job) error {
	txEnvelope := job.TxRequestEnvelope(map[string]string{
		authutils.TenantIDHeader: job.TenantID,
		authutils.UsernameHeader: job.OwnerID,
	})

	evlp, err := txEnvelope.Envelope()
	if err != nil {
		return errors.InvalidParameterError("failed to craft envelope (%s)", err.Error())
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
	}

	if partitionKey := evlp.PartitionKey(); partitionKey != "" {
		msg.Key = sarama.StringEncoder(partitionKey)
	}

	err = saramaencoding.Marshal(txEnvelope, msg)
	if err != nil {
		errMsg := "failed to encode envelope request"
		return errors.EncodingError(errMsg)
	}

	// Send message
	_, _, err = kafkaProducer.SendMessage(msg)
	if err != nil {
		errMsg := "could not produce kafka message"
		return errors.KafkaConnectionError(errMsg)
	}

	return nil
}
