package testutils

import (
	"context"
	encoding "encoding/json"
	"fmt"

	"github.com/consensys/orchestrate/src/api"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
)

const messageListenerComponent = "test.consumer.listener"

type MessageListener struct {
	chanRegistry *ChanRegistry
	topics       *api.TopicConfig
	cancel       context.CancelFunc
	err          error
	logger       *log.Logger
}

func NewMessageListener(chanRegistry *ChanRegistry, topics *api.TopicConfig) *MessageListener {
	return &MessageListener{
		chanRegistry: chanRegistry,
		topics:       topics,
		logger:       log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (l *MessageListener) Setup(session sarama.ConsumerGroupSession) error {
	l.logger.WithContext(session.Context()).
		WithField("kafka.generation_id", session.GenerationID()).
		WithField("kafka.member_id", session.MemberID()).
		WithField("claims", session.Claims()).
		Info("ready to consume messages")

	return nil
}

func (l *MessageListener) Cleanup(session sarama.ConsumerGroupSession) error {
	logger := l.logger.WithContext(session.Context())
	logger.Info("all claims consumed")
	if l.cancel != nil {
		logger.Debug("canceling context")
		l.cancel()
	}

	return l.err
}

func (l *MessageListener) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var ctx context.Context
	ctx, l.cancel = context.WithCancel(session.Context())
	l.err = l.consumeClaimLoop(ctx, session, claim)
	return l.err
}

// nolint
func (l *MessageListener) consumeClaimLoop(ctx context.Context, session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger := l.logger.WithContext(ctx)
	logger.Info("started consuming test claims loop")

	for {
		select {
		case <-ctx.Done():
			logger.WithField("reason", ctx.Err().Error()).Info("gracefully stopping message listener...")
			return nil
		case msg, ok := <-claim.Messages():
			// Input channel has been close so we leave the loop
			if !ok {
				return nil
			}

			msgID, msgObj, err := l.decodeMessage(msg)
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
				Debug("message has been processed")

			session.MarkMessage(msg, "")
			session.Commit()
		}
	}
}

func (l *MessageListener) decodeMessage(msg *sarama.ConsumerMessage) (id string, msgObj interface{}, err error) {
	switch msg.Topic {
	case l.topics.Sender:
		job := &entities.Job{}
		err := encoding.Unmarshal(msg.Value, job)
		if err != nil {
			errMessage := "failed to decode Job message"
			return "", nil, errors.EncodingError(errMessage).ExtendComponent(messageListenerComponent)
		}
		return job.UUID, job, nil
	default:
		return "", nil, errors.InvalidParameterError("message topic cannot be handle %s", msg.Topic)
	}
}

func keyGenOf(key, topic string) string {
	return fmt.Sprintf("%s/%s", topic, key)
}
