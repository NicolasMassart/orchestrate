package testutils

import (
	"context"
	"fmt"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/kafka/testutils"
	"github.com/consensys/orchestrate/src/infra/messenger"
	kafka2 "github.com/consensys/orchestrate/src/infra/messenger/kafka"
	"github.com/consensys/orchestrate/src/infra/notifier/types"
)

const messageConsumerTrackerComponent = "tests.consumer.tracker"

type NotifierConsumerTracker struct {
	consumer messenger.Consumer
	tracker  *testutils.ConsumerTracker
	logger   *log.Logger
}

func NewNotifierConsumerTracker(cfg *sarama.Config, topics []string) (*NotifierConsumerTracker, error) {
	chanRegistry := utils.NewChanRegistry()
	msgConsumerHandler := newNotifierConsumerHandler(chanRegistry)
	msgConsumer, err := kafka2.NewMessageConsumer(cfg, topics, msgConsumerHandler)
	if err != nil {
		return nil, err
	}

	logger := log.NewLogger().SetComponent(messageConsumerTrackerComponent)

	return &NotifierConsumerTracker{
		consumer: msgConsumer,
		tracker:  testutils.NewConsumerTracker(chanRegistry, keyGenOf, logger),
		logger:   logger,
	}, nil
}

func (m *NotifierConsumerTracker) Consume(ctx context.Context) error {
	return m.consumer.Consume(ctx)
}

func (m *NotifierConsumerTracker) Close() error {
	return m.consumer.Close()
}

func (m *NotifierConsumerTracker) WaitForTxMinedNotification(ctx context.Context, uuid, topic string, timeout time.Duration) (*types.Notification, error) {
	msg, err := m.tracker.WaitForMessage(ctx, uuid, topic, timeout)
	if err != nil {
		return nil, err
	}
	notification, ok := msg.(*types.Notification)
	if !ok {
		return nil, fmt.Errorf("invalid message format")
	}

	if notification.Type != types.TransactionMinedMessage {
		return nil, fmt.Errorf("invalid notification type %v", notification.Type)
	}

	return notification, nil
}

func (m *NotifierConsumerTracker) WaitForTxFailedNotification(ctx context.Context, uuid, topic string, timeout time.Duration) (*types.Notification, error) {
	msg, err := m.tracker.WaitForMessage(ctx, uuid, topic, timeout)
	if err != nil {
		return nil, err
	}
	notification, ok := msg.(*types.Notification)
	if !ok {
		return nil, fmt.Errorf("invalid message format")
	}

	if notification.Type != types.TransactionFailedMessage {
		return nil, fmt.Errorf("invalid notification type %v", notification.Type)
	}

	return notification, nil
}
