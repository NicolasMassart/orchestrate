package notifications

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	encoding "github.com/consensys/orchestrate/pkg/encoding/proto"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
	pkgsarama "github.com/consensys/orchestrate/src/infra/broker/sarama"
)

const notifyFailedJobComponent = "use-cases.notify-failed-job"

type notifyFailedJobUseCase struct {
	db            store.DB
	kafkaProducer sarama.SyncProducer
	topicsCfg     *pkgsarama.KafkaTopicConfig
	logger        *log.Logger
}

func NewNotifyFailedJobUseCase(
	db store.DB,
	kafkaProducer sarama.SyncProducer,
	topicsCfg *pkgsarama.KafkaTopicConfig,
) usecases.NotifyFailedJob {
	return &notifyFailedJobUseCase{
		db:            db,
		kafkaProducer: kafkaProducer,
		topicsCfg:     topicsCfg,
		logger:        log.NewLogger().SetComponent(notifyFailedJobComponent),
	}
}

func (uc *notifyFailedJobUseCase) Execute(_ context.Context, job *entities.Job, errMsg string) error {
	logger := uc.logger.WithField("job", job.UUID)

	msg := job.TxResponse(errors.FromError(fmt.Errorf(errMsg)))
	b, err := encoding.Marshal(msg)
	if err != nil {
		errMsg := "failed to encode failed job response"
		logger.WithError(err).Error(errMsg)
		return errors.EncodingError(errMsg)
	}

	// Send message
	_, _, err = uc.kafkaProducer.SendMessage(&sarama.ProducerMessage{
		Topic: uc.topicsCfg.Recover,
		Key:   sarama.StringEncoder(job.ChainUUID),
		Value: sarama.ByteEncoder(b),
	})
	if err != nil {
		errMsg := "could not send kafka notification"
		logger.WithError(err).Error(errMsg)
		return errors.KafkaConnectionError(errMsg)
	}

	logger.WithField("msg_id", msg.Id).
		WithField("topic", uc.topicsCfg.Recover).
		Debug("failed job notification was sent successfully")

	return nil
}
