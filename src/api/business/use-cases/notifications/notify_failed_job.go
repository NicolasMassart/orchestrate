package notifications

import (
	"context"
	"fmt"

	"github.com/consensys/orchestrate/src/infra/kafka"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/kafka/proto"
	broker "github.com/consensys/orchestrate/src/infra/kafka/sarama"
)

const notifyFailedJobComponent = "use-cases.notify-failed-job"

type notifyFailedJobUseCase struct {
	db            store.DB
	kafkaProducer kafka.Producer
	topicsCfg     *broker.TopicConfig
	logger        *log.Logger
}

func NewNotifyFailedJobUseCase(
	db store.DB,
	kafkaProducer kafka.Producer,
	topicsCfg *broker.TopicConfig,
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

	msg := proto.NewTxResponse(job, errors.FromError(fmt.Errorf(errMsg)))
	// Send message
	err := uc.kafkaProducer.SendTxResponse(uc.topicsCfg.Recover, msg)
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
