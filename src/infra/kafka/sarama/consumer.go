package sarama

import (
	"context"

	"github.com/consensys/orchestrate/src/infra/kafka"

	"github.com/Shopify/sarama"
)

type Consumer struct {
	cGroup sarama.ConsumerGroup
	client kafka.Client
}

var _ kafka.Consumer = &Consumer{}

func NewConsumerGroup(cfg *Config) (*Consumer, error) {
	client, err := NewClient(cfg, cfg.URLs)
	if err != nil {
		return nil, err
	}

	cGroup, err := NewConsumerGroupFromClient(client, cfg.GroupName)
	if err != nil {
		return nil, err
	}

	return &Consumer{cGroup: cGroup, client: client}, nil
}

func (c *Consumer) Client() kafka.Client {
	return c.client
}

// Consume start consuming using global ConsumerGroup
func (c *Consumer) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
consumeLoop:
	for {
		select {
		case <-ctx.Done():
			break consumeLoop
		default:
			err := c.cGroup.Consume(ctx, topics, handler)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Consumer) Close() error {
	return c.cGroup.Close()
}
