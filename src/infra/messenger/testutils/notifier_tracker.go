package testutils

import (
	"context"
	"fmt"
	"time"

	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/src/infra/messenger/types"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/infra/messenger"
	"github.com/consensys/orchestrate/src/infra/messenger/kafka"
)

const notifierMessageConsumerTrackerComponent = "tests.consumer.tracker"

type NotifierConsumerTracker struct {
	consumer messenger.Consumer
	tracker  *ConsumerTracker
	logger   *log.Logger
}

func NewNotifierConsumerTracker(cfg *kafka.Config, topics []string) (*NotifierConsumerTracker, error) {
	chanRegistry := utils.NewChanRegistry()
	msgConsumerHandler := newNotifierConsumerHandler(chanRegistry)
	msgConsumer, err := kafka.NewMessageConsumer(cfg, topics, msgConsumerHandler)
	if err != nil {
		return nil, err
	}

	logger := log.NewLogger().SetComponent(notifierMessageConsumerTrackerComponent)

	return &NotifierConsumerTracker{
		consumer: msgConsumer,
		tracker:  NewConsumerTracker(chanRegistry, keyGenOf, logger),
		logger:   logger,
	}, nil
}

func (m *NotifierConsumerTracker) Consume(ctx context.Context) error {
	return m.consumer.Consume(ctx)
}

func (m *NotifierConsumerTracker) Close() error {
	return m.consumer.Close()
}

func (m *NotifierConsumerTracker) WaitForTxMinedNotification(ctx context.Context, id, topic string, timeout time.Duration) (*types.NotificationResponse, error) {
	msg, err := m.tracker.WaitForMessage(ctx, id, topic, timeout)
	msgID := fmt.Sprintf("Message '%s' on topic '%s'", id, topic)
	if err != nil {
		return nil, fmt.Errorf("%s. %s", err.Error(), msgID)
	}
	notification, ok := msg.(*types.NotificationResponse)
	if !ok {
		return nil, fmt.Errorf("invalid message format. %s", msgID)
	}

	if notification.Type != string(entities.NotificationTypeTxMined) {
		return nil, fmt.Errorf("invalid notification type %v. %s", notification.Type, msgID)
	}

	return notification, nil
}

func (m *NotifierConsumerTracker) WaitForTxFailedNotification(ctx context.Context, uuid, topic string, timeout time.Duration) (*types.NotificationResponse, error) {
	msg, err := m.tracker.WaitForMessage(ctx, uuid, topic, timeout)
	if err != nil {
		return nil, err
	}
	notification, ok := msg.(*types.NotificationResponse)
	if !ok {
		return nil, fmt.Errorf("invalid message format")
	}

	if notification.Type != string(entities.NotificationTypeTxFailed) {
		return nil, fmt.Errorf("invalid notification type %v", notification.Type)
	}

	return notification, nil
}
