package testutils

import (
	"context"
	"fmt"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/kafka"
	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/push_notification/client"
)

const messageConsumerTrackerComponent = "tests.consumer.tracker"

type InternalConsumerTracker struct {
	consumerTracker
}

func NewInternalConsumerTracker(cfg *sarama.Config) (*InternalConsumerTracker, error) {
	consumer, err := sarama.NewConsumer(cfg)
	if err != nil {
		return nil, err
	}

	chanRegistry := newChanRegistry()
	listener := newInternalMessageListener(chanRegistry)

	return &InternalConsumerTracker{consumerTracker{
		consumer:     consumer,
		chanRegistry: chanRegistry,
		listener:     listener,
		logger:       log.NewLogger().SetComponent(messageConsumerTrackerComponent),
	}}, nil
}

func (m *InternalConsumerTracker) WaitForJob(ctx context.Context, jobUUID, topic string, timeout time.Duration) (*entities.Job, error) {
	msg, err := m.waitForMessage(ctx, jobUUID, topic, timeout)
	if err != nil {
		return nil, err
	}
	msgJob, ok := msg.(*entities.Job)
	if !ok {
		return nil, fmt.Errorf("invalid message type")
	}
	return msgJob, nil
}

type ExternalConsumerTracker struct {
	consumerTracker
}

func NewExternalConsumerTracker(cfg *sarama.Config) (*ExternalConsumerTracker, error) {
	consumer, err := sarama.NewConsumer(cfg)
	if err != nil {
		return nil, err
	}

	chanRegistry := newChanRegistry()
	listener := newNotificationMessageListener(chanRegistry)

	return &ExternalConsumerTracker{consumerTracker{
		consumer:     consumer,
		chanRegistry: chanRegistry,
		listener:     listener,
		logger:       log.NewLogger().SetComponent(messageConsumerTrackerComponent),
	}}, nil
}

func (m *ExternalConsumerTracker) WaitForTxMinedNotification(ctx context.Context, uuid, topic string, timeout time.Duration) (*client.Notification, error) {
	msg, err := m.waitForMessage(ctx, uuid, topic, timeout)
	if err != nil {
		return nil, err
	}
	notification, ok := msg.(*client.Notification)
	if !ok {
		return nil, fmt.Errorf("invalid message format")
	}

	if notification.Type != client.TransactionMinedMessage {
		return nil, fmt.Errorf("invalid notification type %v", notification.Type)
	}

	return notification, nil
}

func (m *ExternalConsumerTracker) WaitForTxFailedNotification(ctx context.Context, uuid, topic string, timeout time.Duration) (*client.Notification, error) {
	msg, err := m.waitForMessage(ctx, uuid, topic, timeout)
	if err != nil {
		return nil, err
	}
	notification, ok := msg.(*client.Notification)
	if !ok {
		return nil, fmt.Errorf("invalid message format")
	}

	if notification.Type != client.TransactionFailedMessage {
		return nil, fmt.Errorf("invalid notification type %v", notification.Type)
	}

	return notification, nil
}

type consumerTracker struct {
	consumer     kafka.Consumer
	chanRegistry *chanRegistry
	listener     *messageListener
	logger       *log.Logger
}

func (m *consumerTracker) Consume(ctx context.Context, topics []string) error {
	return m.consumer.Consume(ctx, topics, m.listener)
}

func (m *consumerTracker) Close() error {
	return m.consumer.Close()
}

func (m *consumerTracker) waitForMessage(ctx context.Context, id, topic string, timeout time.Duration) (interface{}, error) {
	logger := m.logger.WithField("id", id).WithField("topic", topic).WithField("timeout", timeout/time.Millisecond)

	logger.Debug("waiting for message...")
	msg, err := m.waitForChanMessage(ctx, id, topic, timeout)
	if err != nil {
		logger.WithError(err).Error("failed to find message")
		return nil, err
	}

	return msg, err
}

func (m *consumerTracker) waitForChanMessage(ctx context.Context, id, topic string, timeout time.Duration) (interface{}, error) {
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
