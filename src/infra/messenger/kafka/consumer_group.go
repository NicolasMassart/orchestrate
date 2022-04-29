package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
)

type ConsumerGroup struct {
	g      sarama.ConsumerGroup
	errors chan error
}

func NewConsumerGroup(cfg *Config) (*ConsumerGroup, error) {
	saramaCfg, err := cfg.ToKafkaConfig()
	if err != nil {
		return nil, err
	}

	g, err := sarama.NewConsumerGroup(cfg.URLs, cfg.GroupName, saramaCfg)
	if err != nil {
		return nil, errors.KafkaConnectionError(err.Error())
	}

	cg := &ConsumerGroup{
		g:      g,
		errors: make(chan error, saramaCfg.ChannelBufferSize),
	}

	// Pipe errors
	go func() {
		for err := range g.Errors() {
			cg.errors <- errors.KafkaConnectionError(err.Error())
		}
	}()

	return cg, nil
}

// Consume implements ConsumerGroup.
func (c *ConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	err := c.g.Consume(ctx, topics, handler)
	if err != nil {
		return errors.KafkaConnectionError(err.Error())
	}
	return nil
}

// Errors implements ConsumerGroup.
func (c *ConsumerGroup) Errors() <-chan error {
	return c.errors
}

func (c *ConsumerGroup) Pause(partitions map[string][]int32) {
	c.g.Pause(partitions)
}

func (c *ConsumerGroup) Resume(partitions map[string][]int32) {
	c.g.Resume(partitions)
}

func (c *ConsumerGroup) PauseAll() {
	c.g.PauseAll()
}

func (c *ConsumerGroup) ResumeAll() {
	c.g.ResumeAll()
}

// Close implements ConsumerGroup.
func (c *ConsumerGroup) Close() error {
	err := c.g.Close()
	if err != nil {
		return errors.KafkaConnectionError(err.Error())
	}
	return nil
}
