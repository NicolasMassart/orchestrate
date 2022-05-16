package listener

import (
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	messenger "github.com/consensys/orchestrate/src/infra/messenger/kafka"
)

const (
	messageListenerComponent = "api.kafka-consumer"
)

func NewMessageConsumer(cfg *kafka.Config,
	topics []string,
	jobHandler *JobHandler,
	subscriptionHandler *SubscriptionHandler,
	notificationHandler *NotificationHandler,
	eventStreamHandler *EventStreamHandler,
) (*messenger.Consumer, error) {
	consumer, err := messenger.NewMessageConsumer(messageListenerComponent, cfg, topics)
	if err != nil {
		return nil, err
	}

	consumer.AppendHandler(UpdateJobMessageType, jobHandler.HandleJobUpdate)
	consumer.AppendHandler(EventLogsMessageType, subscriptionHandler.HandleEventLogs)
	consumer.AppendHandler(AckNotificationMessageType, notificationHandler.HandleNotificationAck)
	consumer.AppendHandler(SuspendEventStreamMessageType, eventStreamHandler.HandleEventStreamSuspend)
	return consumer, nil
}
