package testutils

import (
	"context"
	"fmt"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/messenger"
	"github.com/consensys/orchestrate/src/infra/messenger/kafka"
)

const jobMessageConsumerTrackerComponent = "tests.consumer.tracker"

type MessengerConsumerTracker struct {
	consumer messenger.Consumer
	tracker  *ConsumerTracker
	logger   *log.Logger
}

func NewMessengerConsumerTracker(cfg *kafka.Config, topics []string) (*MessengerConsumerTracker, error) {
	chanRegistry := utils.NewChanRegistry()
	msgConsumerHandler := newMessageConsumerHandler(chanRegistry)
	msgConsumer, err := kafka.NewMessageConsumer(cfg, topics, msgConsumerHandler)
	if err != nil {
		return nil, err
	}

	logger := log.NewLogger().SetComponent(jobMessageConsumerTrackerComponent)

	return &MessengerConsumerTracker{
		consumer: msgConsumer,
		tracker:  NewConsumerTracker(chanRegistry, keyGenOf, logger),
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
