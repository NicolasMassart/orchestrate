package testutils

import (
	"context"
	encoding "encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/push_notification/client"
)

const messageListenerComponent = "test.consumer.listener"

type decodeMessageFunc func(msg *sarama.ConsumerMessage) (id string, msgObj interface{}, err error)

type messageListener struct {
	chanRegistry      *chanRegistry
	decodeMessageFunc decodeMessageFunc
	cancel            context.CancelFunc
	err               error
	logger            *log.Logger
}

func newInternalMessageListener(chanRegistry *chanRegistry) *messageListener {
	return &messageListener{
		chanRegistry:      chanRegistry,
		decodeMessageFunc: decodeInternalMessage,
		logger:            log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func newNotificationMessageListener(chanRegistry *chanRegistry) *messageListener {
	return &messageListener{
		chanRegistry:      chanRegistry,
		decodeMessageFunc: decodeNotificationMessage,
		logger:            log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (l *messageListener) Setup(session sarama.ConsumerGroupSession) error {
	l.logger.WithContext(session.Context()).
		WithField("kafka.generation_id", session.GenerationID()).
		WithField("kafka.member_id", session.MemberID()).
		WithField("claims", session.Claims()).
		Debug("ready to consume messages")

	return nil
}

func (l *messageListener) Cleanup(session sarama.ConsumerGroupSession) error {
	logger := l.logger.WithContext(session.Context())
	logger.Info("all claims consumed")
	if l.cancel != nil {
		logger.Debug("canceling context")
		l.cancel()
	}

	return l.err
}

func (l *messageListener) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var ctx context.Context
	ctx, l.cancel = context.WithCancel(session.Context())
	l.err = l.consumeClaimLoop(ctx, session, claim)
	return l.err
}

// nolint
func (l *messageListener) consumeClaimLoop(ctx context.Context, session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger := l.logger.WithContext(ctx)
	logger.Debug("started consuming test claims loop")

	for {
		select {
		case <-ctx.Done():
			logger.WithField("reason", ctx.Err().Error()).Debug("gracefully stopping message listener...")
			return nil
		case msg, ok := <-claim.Messages():
			// Input channel has been close so we leave the loop
			if !ok {
				return nil
			}

			msgID, msgObj, err := l.decodeMessageFunc(msg)
			if err != nil {
				logger.WithError(err).Error("error decoding message", msg)
				session.MarkMessage(msg, "")
				continue
			}

			msgKey := keyGenOf(msgID, msg.Topic)
			if !l.chanRegistry.HasChan(msgKey) {
				l.chanRegistry.Register(msgKey, make(chan interface{}, 1))
			}

			// Dispatch envelope
			err = l.chanRegistry.Send(msgKey, msgObj)
			if err != nil {
				logger.WithError(err).Error("message dispatched with errors")
				continue
			}

			logger.WithField("id", msgID).WithField("topic", msg.Topic).
				Info("message has been processed")

			session.MarkMessage(msg, "")
			session.Commit()
		}
	}
}

func decodeInternalMessage(msg *sarama.ConsumerMessage) (id string, msgObj interface{}, err error) {
	job := &entities.Job{}
	err = encoding.Unmarshal(msg.Value, job)
	if err != nil {
		errMessage := "failed to decode Job message"
		return "", nil, errors.EncodingError(errMessage).ExtendComponent(messageListenerComponent)
	}
	return job.UUID, job, nil
}

func decodeNotificationMessage(msg *sarama.ConsumerMessage) (id string, msgObj interface{}, err error) {
	notification := &client.Notification{}
	err = encoding.Unmarshal(msg.Value, notification)
	if err != nil {
		errMessage := "failed to decode Notification message"
		return "", nil, errors.EncodingError(errMessage).ExtendComponent(messageListenerComponent)
	}
	return notification.UUID, notification, nil
}

func keyGenOf(key, topic string) string {
	return fmt.Sprintf("%s/%s", topic, key)
}
