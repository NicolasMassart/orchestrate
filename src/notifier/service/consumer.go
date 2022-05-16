package service

import (
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	messenger "github.com/consensys/orchestrate/src/infra/messenger/kafka"
)

const (
	messageListenerComponent = "notifier.kafka-consumer"
)

func NewMessageConsumer(cfg *kafka.Config,
	topics []string,
	transactionHandler *TransactionHandler,
	subscriptionHandler *SubscriptionHandler,
) (*messenger.Consumer, error) {
	consumer, err := messenger.NewMessageConsumer(messageListenerComponent, cfg, topics)
	if err != nil {
		return nil, err
	}

	consumer.AppendHandler(TransactionMessageType, transactionHandler.HandleTransactionReq)
	consumer.AppendHandler(ContractEventMessageType, subscriptionHandler.HandleContractEventReq)
	return consumer, nil
}
