package service

import (
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	messenger "github.com/consensys/orchestrate/src/infra/messenger/kafka"
)

const (
	messageListenerComponent = "service.kafka-consumer"
)

func NewMessageConsumer(cfg *kafka.Config, topics []string, jobHandler *JobHandler) (*messenger.Consumer, error) {
	consumer, err := messenger.NewMessageConsumer(messageListenerComponent, cfg, topics)
	if err != nil {
		return nil, err
	}

	consumer.AppendHandler(StartedJobMessageType, jobHandler.HandleStartedJob)

	return consumer, nil
}
