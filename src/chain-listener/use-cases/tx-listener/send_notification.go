package txlistener

import (
	"context"

	"github.com/Shopify/sarama"
	encoding "github.com/consensys/orchestrate/pkg/encoding/sarama"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/entities"
)

const sendNotificationComponent = "chain-listener.use-case.tx-listener.send-notification"

type sendNotificationUseCase struct {
	producer sarama.SyncProducer
	topic    string
	logger   *log.Logger
}

func SendNotificationUseCase(producer sarama.SyncProducer, topic string, logger *log.Logger) usecases.SendNotification {
	return &sendNotificationUseCase{
		producer: producer,
		topic:    topic,
		logger:   logger.SetComponent(sendNotificationComponent),
	}
}

func (uc *sendNotificationUseCase) Execute(_ context.Context, job *entities.Job) error {
	logger := uc.logger.WithField("chain", job.ChainUUID).WithField("job", job.UUID).WithField("tx_hash", job.Transaction.Hash)
	logger.Debug("sending job notification")

	msg := &sarama.ProducerMessage{}
	err := encoding.Marshal(job.TxResponse(), msg)
	if err != nil {
		return err
	}

	msg.Topic = uc.topic
	msg.Key = sarama.StringEncoder(job.ChainUUID)
	_, _, err = uc.producer.SendMessage(msg)
	if err != nil {
		logger.WithError(err).Errorf("failed to produce message")
		return err
	}

	logger.Info("notification has been sent")
	return nil
}
