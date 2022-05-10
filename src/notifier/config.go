package notifier

import (
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
)

type Config struct {
	Kafka         *kafka.Config
	TopicAPI      string
	ConsumerTopic string
	MaxRetries    int
}
