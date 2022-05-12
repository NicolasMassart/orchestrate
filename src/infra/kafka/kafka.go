package kafka

import (
	"context"

	"github.com/Shopify/sarama"
)

//go:generate mockgen -source=kafka.go -destination=mocks/kafka.go -package=mocks

type ConsumerGroup interface {
	Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error
	Errors() <-chan error
	Close() error
	Checker() error
}

type Producer interface {
	Send(body interface{}, topic string, partitionKey string, headers map[string]interface{}) error
	Close() error
	Checker() error
}
