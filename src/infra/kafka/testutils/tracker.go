package testutils

import (
	"context"
	"fmt"
	"time"

	"github.com/consensys/orchestrate/src/api"
	"github.com/consensys/orchestrate/src/infra/push_notification/client"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/kafka"
	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
)

const messageConsumerTrackerComponent = "tests.consumer.tracker"

type ConsumerTracker struct {
	consumer     kafka.Consumer
	chanRegistry *ChanRegistry
	topics       *api.TopicConfig
	logger       *log.Logger
}

func NewConsumerTracker(cfg *sarama.Config, topics *api.TopicConfig) (*ConsumerTracker, error) {
	consumer, err := sarama.NewConsumer(cfg)
	if err != nil {
		return nil, err
	}

	chanRegistry := NewChanRegistry()

	return &ConsumerTracker{
		consumer:     consumer,
		chanRegistry: chanRegistry,
		topics:       topics,
		logger:       log.NewLogger().SetComponent(messageConsumerTrackerComponent),
	}, nil
}
func (m *ConsumerTracker) Consume(ctx context.Context, topics []string) error {
	listener := NewMessageListener(m.chanRegistry, m.topics)
	return m.consumer.Consume(ctx, topics, listener)
}

func (m *ConsumerTracker) Close() error {
	return m.consumer.Close()
}

func (m *ConsumerTracker) WaitForTxResponseInTopic(ctx context.Context, id, topic string, timeout time.Duration) (*client.TxResponse, error) {
	logger := m.logger.WithField("id", id).WithField("topic", topic).WithField("timeout", timeout/time.Millisecond)

	logger.Debug("waiting for txResponse...")
	msg, err := m.waitForMessage(ctx, id, topic, timeout)
	if err != nil {
		logger.WithError(err).Error("failed to find txResponse")
		return nil, err
	}

	txResponse, ok := msg.(*client.TxResponse)
	if !ok {
		err := fmt.Errorf("failed to decode txResponse")
		logger.Error(err.Error())
		return nil, err
	}

	return txResponse, nil
}

func (m *ConsumerTracker) WaitForJob(ctx context.Context, id, topic string, timeout time.Duration) (*entities.Job, error) {
	logger := m.logger.WithField("id", id).WithField("topic", topic).WithField("timeout", timeout/time.Millisecond)

	logger.Debug("waiting for job...")
	msg, err := m.waitForMessage(ctx, id, topic, timeout)
	if err != nil {
		logger.WithError(err).Error("failed to find job")
		return nil, err
	}

	job, ok := msg.(*entities.Job)
	if !ok {
		err := fmt.Errorf("failed to decode job")
		logger.Error(err.Error())
		return nil, err
	}

	return job, nil
}

func (m *ConsumerTracker) waitForMessage(ctx context.Context, id, topic string, timeout time.Duration) (interface{}, error) {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var ch = make(chan interface{}, 1)
	go func(chx chan interface{}) {
		msgKey := keyGenOf(id, topic)
		if !m.chanRegistry.HasChan(msgKey) {
			m.chanRegistry.Register(msgKey, make(chan interface{}, 1))
		}

		e := <-m.chanRegistry.GetChan(msgKey)
		chx <- e
	}(ch)

	select {
	case e := <-ch:
		return e, nil
	case <-cctx.Done():
		return nil, cctx.Err()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
