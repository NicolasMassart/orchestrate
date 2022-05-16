package trackers

import (
	"context"
	"encoding/json"
	"github.com/consensys/orchestrate/src/api/service/listener"
	"github.com/consensys/orchestrate/src/api/service/types"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/messenger"
	messengerkafka "github.com/consensys/orchestrate/src/infra/messenger/kafka"
	notifier2 "github.com/consensys/orchestrate/src/notifier/service"
	notifierTypes "github.com/consensys/orchestrate/src/notifier/service/types"
	txlistener "github.com/consensys/orchestrate/src/tx-listener/service"
	txlistenerTypes "github.com/consensys/orchestrate/src/tx-listener/service/types"
	txsender "github.com/consensys/orchestrate/src/tx-sender/service"
	txsenderTypes "github.com/consensys/orchestrate/src/tx-sender/service/types"
)

const testConsumerTrackerComponent = "tests.messenger.tracker"
const invalidTypeErr = "invalid message type"

type KeyGenFunc func(req interface{}) string

type MessengerConsumerTracker struct {
	consumer     messenger.Consumer
	tracker      *ConsumerTracker
	chanRegistry *utils.ChanRegistry
	logger       *log.Logger
}

func NewMessengerConsumerTracker(cfg kafka.Config, topics []string) (*MessengerConsumerTracker, error) {
	chanRegistry := utils.NewChanRegistry()
	cfg.GroupName = testConsumerTrackerComponent + utils.RandString(5)
	msgConsumer, err := messengerkafka.NewMessageConsumer(testConsumerTrackerComponent, &cfg, topics)
	if err != nil {
		return nil, err
	}

	logger := log.NewLogger().SetComponent(testConsumerTrackerComponent)
	msg := &MessengerConsumerTracker{
		consumer:     msgConsumer,
		chanRegistry: chanRegistry,
		tracker:      NewConsumerTracker(chanRegistry, logger),
		logger:       logger,
	}

	msg.trackMessageType(txsender.StartedJobMessageType, &txsenderTypes.StartedJobReq{}, func(req interface{}) string {
		return req.(*txsenderTypes.StartedJobReq).Job.UUID
	})
	msg.trackMessageType(txlistener.PendingJobMessageType, &txlistenerTypes.PendingJobMessageRequest{}, func(req interface{}) string {
		return req.(*txlistenerTypes.PendingJobMessageRequest).Job.UUID
	})
	msg.trackMessageType(notifier2.TransactionMessageType, &notifierTypes.TransactionMessageRequest{}, func(req interface{}) string {
		return req.(*notifierTypes.TransactionMessageRequest).Notification.Job.UUID
	})
	msg.trackMessageType(listener.SuspendEventStreamMessageType, &types.SuspendEventStreamRequestMessage{}, func(req interface{}) string {
		return req.(*types.SuspendEventStreamRequestMessage).UUID
	})
	msg.trackMessageType(listener.AckNotificationMessageType, &types.AckNotificationRequestMessage{}, func(req interface{}) string {
		return req.(*types.AckNotificationRequestMessage).UUID
	})

	return msg, nil
}

func (m *MessengerConsumerTracker) trackMessageType(msgType messenger.ConsumerRequestMessageType, req interface{}, keyGenFunc KeyGenFunc) {
	m.consumer.AppendHandler(msgType, m.trackMessageHandle(req, msgType, keyGenFunc))
}

func (m *MessengerConsumerTracker) WaitForStartedJobMessage(ctx context.Context, msgId string, timeout time.Duration) (*txsenderTypes.StartedJobReq, error) {
	msg, err := m.tracker.WaitForMessage(ctx, keyGenOf(msgId, string(txsender.StartedJobMessageType)), timeout)
	if err != nil {
		return nil, err
	}
	req, ok := msg.(*txsenderTypes.StartedJobReq)
	if !ok {
		return nil, errors.EncodingError(invalidTypeErr)
	}
	return req, nil
}

func (m *MessengerConsumerTracker) WaitForPendingJobMessage(ctx context.Context, msgId string, timeout time.Duration) (*txlistenerTypes.PendingJobMessageRequest, error) {
	msg, err := m.tracker.WaitForMessage(ctx, keyGenOf(msgId, string(txlistener.PendingJobMessageType)), timeout)
	if err != nil {
		return nil, err
	}
	req, ok := msg.(*txlistenerTypes.PendingJobMessageRequest)
	if !ok {
		return nil, errors.EncodingError(invalidTypeErr)
	}
	return req, nil
}

func (m *MessengerConsumerTracker) WaitForTransactionNotificationMessage(ctx context.Context, msgId string, timeout time.Duration) (*notifierTypes.TransactionMessageRequest, error) {
	msg, err := m.tracker.WaitForMessage(ctx, keyGenOf(msgId, string(notifier2.TransactionMessageType)), timeout)
	if err != nil {
		return nil, err
	}
	req, ok := msg.(*notifierTypes.TransactionMessageRequest)
	if !ok {
		return nil, errors.EncodingError(invalidTypeErr)
	}
	return req, nil
}

func (m *MessengerConsumerTracker) WaitForSuspendEventStream(ctx context.Context, msgId string, timeout time.Duration) (*types.SuspendEventStreamRequestMessage, error) {
	msg, err := m.tracker.WaitForMessage(ctx, keyGenOf(msgId, listener.SuspendEventStreamMessageType), timeout)
	if err != nil {
		return nil, err
	}

	req, ok := msg.(*types.SuspendEventStreamRequestMessage)
	if !ok {
		return nil, errors.EncodingError(invalidTypeErr)
	}

	return req, nil
}

func (m *MessengerConsumerTracker) WaitForAckNotif(ctx context.Context, msgId string, timeout time.Duration) (*types.AckNotificationRequestMessage, error) {
	msg, err := m.tracker.WaitForMessage(ctx, keyGenOf(msgId, listener.AckNotificationMessageType), timeout)
	if err != nil {
		return nil, err
	}

	req, ok := msg.(*types.AckNotificationRequestMessage)
	if !ok {
		return nil, errors.EncodingError(invalidTypeErr)
	}

	return req, nil
}

func (m *MessengerConsumerTracker) Consume(ctx context.Context) error {
	return m.consumer.Consume(ctx)
}

func (m *MessengerConsumerTracker) Close() error {
	return m.consumer.Close()
}

func (m *MessengerConsumerTracker) trackMessageHandle(req interface{}, msgType messenger.ConsumerRequestMessageType, keyGenFunc KeyGenFunc) messenger.MessageHandler {
	return func(_ context.Context, rawReq []byte) error {
		err := json.Unmarshal(rawReq, req)

		msgId := keyGenOf(keyGenFunc(req), string(msgType))
		if !m.chanRegistry.HasChan(msgId) {
			m.chanRegistry.Register(msgId, make(chan interface{}, 1))
		}

		logger := m.logger.WithField("msg", msgId)
		// Dispatch envelope
		err = m.chanRegistry.Send(msgId, req)
		if err != nil {
			logger.WithError(err).Error("message dispatched with errors")
			return err
		}

		logger.Info("message has been processed")
		return nil
	}
}
