package notifier

import (
	"github.com/consensys/orchestrate/pkg/sdk/messenger"
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
)

type Config struct {
	Kafka         *kafka.Config
	Messenger     *messenger.Config
	ConsumerTopic string
	MaxRetries    int
}
