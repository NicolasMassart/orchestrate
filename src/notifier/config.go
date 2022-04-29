package notifier

import (
	"github.com/consensys/orchestrate/src/infra/messenger/kafka"
)

type Config struct {
	Kafka         *kafka.Config
	ConsumerTopic string
	NConsumer     int
	MaxRetries    int
}
