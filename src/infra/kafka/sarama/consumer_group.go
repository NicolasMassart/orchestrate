package sarama

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
)

type consumerGroup struct {
	g      sarama.ConsumerGroup
	errors chan error
}

func newConsumerGroupFromClient(groupID string, client sarama.Client) (*consumerGroup, error) {
	g, err := sarama.NewConsumerGroupFromClient(groupID, client)
	if err != nil {
		return nil, errors.KafkaConnectionError(err.Error())
	}

	cg := &consumerGroup{
		g:      g,
		errors: make(chan error, client.Config().ChannelBufferSize),
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
func (c *consumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	err := c.g.Consume(ctx, topics, handler)
	if err != nil {
		return errors.KafkaConnectionError(err.Error())
	}
	return nil
}

// Errors implements ConsumerGroup.
func (c *consumerGroup) Errors() <-chan error {
	return c.errors
}

func (c *consumerGroup) Pause(partitions map[string][]int32) {
	c.g.Pause(partitions)
}

func (c *consumerGroup) Resume(partitions map[string][]int32) {
	c.g.Resume(partitions)
}

func (c *consumerGroup) PauseAll() {
	c.g.PauseAll()
}

func (c *consumerGroup) ResumeAll() {
	c.g.ResumeAll()
}

// Close implements ConsumerGroup.
func (c *consumerGroup) Close() error {
	err := c.g.Close()
	if err != nil {
		return errors.KafkaConnectionError(err.Error())
	}
	return nil
}
