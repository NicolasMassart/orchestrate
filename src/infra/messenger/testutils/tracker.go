package testutils

import (
	"context"
	"fmt"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/kafka/testutils"
	"github.com/consensys/orchestrate/src/infra/messenger"
	kafka2 "github.com/consensys/orchestrate/src/infra/messenger/kafka"
)

const messageConsumerTrackerComponent = "tests.consumer.tracker"

type MessengerConsumerTracker struct {
	consumer messenger.Consumer
	tracker  *testutils.ConsumerTracker
	logger   *log.Logger
}

func NewMessengerConsumerTracker(cfg *sarama.Config, topics []string) (*MessengerConsumerTracker, error) {
	chanRegistry := utils.NewChanRegistry()
	msgConsumerHandler := newMessageConsumerHandler(chanRegistry)
	msgConsumer, err := kafka2.NewMessageConsumer(cfg, topics, msgConsumerHandler)
	if err != nil {
		return nil, err
	}

	logger := log.NewLogger().SetComponent(messageConsumerTrackerComponent)

	return &MessengerConsumerTracker{
		consumer: msgConsumer,
		tracker:  testutils.NewConsumerTracker(chanRegistry, keyGenOf, logger),
		logger:   logger,
	}, nil
}

func (m *MessengerConsumerTracker) Consume(ctx context.Context) error {
	return m.consumer.Consume(ctx)
}

func (m *MessengerConsumerTracker) Close() error {
	return m.consumer.Close()
}

func (m *MessengerConsumerTracker) WaitForJob(ctx context.Context, jobUUID, topic string, timeout time.Duration) (*entities.Job, error) {
	msg, err := m.tracker.WaitForMessage(ctx, jobUUID, topic, timeout)
	if err != nil {
		return nil, err
	}
	msgJob, ok := msg.(*entities.Job)
	if !ok {
		return nil, fmt.Errorf("invalid message type")
	}
	return msgJob, nil
}
