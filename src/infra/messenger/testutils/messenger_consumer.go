package testutils

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	messenger "github.com/consensys/orchestrate/src/infra/messenger/kafka"
)

const (
	jobMessageListenerComponent = "test.service.kafka-job-consumer"
)

type messageConsumerHandler struct {
	chanRegistry *utils.ChanRegistry
	logger       *log.Logger
}

var _ messenger.ConsumerMessageHandler = &messageConsumerHandler{}

func newMessageConsumerHandler(chanRegistry *utils.ChanRegistry) *messageConsumerHandler {
	return &messageConsumerHandler{
		chanRegistry: chanRegistry,
		logger:       log.NewLogger().SetComponent(jobMessageListenerComponent),
	}
}

func (mch *messageConsumerHandler) ProcessMsg(_ context.Context, rawMsg *sarama.ConsumerMessage, decodedMsg interface{}) error {
	job := decodedMsg.(*entities.Job)
	logger := mch.logger.WithField("job", job.UUID).WithField("schedule", job.ScheduleUUID).WithField("topic", rawMsg.Topic)
	msgKey := keyGenOf(job.UUID, rawMsg.Topic)
	if !mch.chanRegistry.HasChan(msgKey) {
		mch.chanRegistry.Register(msgKey, make(chan interface{}, 1))
	}

	// Dispatch envelope
	err := mch.chanRegistry.Send(msgKey, job)
	if err != nil {
		logger.WithError(err).Error("message dispatched with errors")
		return err
	}

	logger.Info("message has been processed")
	return nil
}

func (mch *messageConsumerHandler) DecodeMessage(rawMsg *sarama.ConsumerMessage) (interface{}, error) {
	return messenger.DecodeJobMessage(rawMsg)
}

func (mch *messageConsumerHandler) ID() string {
	return jobMessageListenerComponent
}
