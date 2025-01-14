package kafka

import (
	"context"
	encoding "encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/kafka"
	kafkasarama "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/messenger"
)

const consumerComponent = "messenger.kafka.consumer"

type Consumer struct {
	consumerGroup       kafka.ConsumerGroup
	handler             map[entities.RequestMessageType]messenger.MessageHandler
	topics              []string
	cancel              context.CancelFunc
	logger              *log.Logger
	disableCommitOnRead bool
	err                 error
}

var _ messenger.Consumer = &Consumer{}

func NewMessageConsumer(id string, cfg *kafkasarama.Config, topics []string) (*Consumer, error) {
	consumerGroup, err := kafkasarama.NewConsumerGroup(cfg)
	if err != nil {
		return nil, err
	}

	consumer := &Consumer{
		consumerGroup:       consumerGroup,
		topics:              topics,
		disableCommitOnRead: cfg.DisableCommitOnRead,
		logger:              log.NewLogger().SetComponent(consumerComponent + "." + id),
		handler:             map[entities.RequestMessageType]messenger.MessageHandler{},
	}

	return consumer, nil
}

func (cl *Consumer) Consume(ctx context.Context) error {
	return cl.consumerGroup.Consume(ctx, cl.topics, cl)
}

func (cl *Consumer) Checker() error {
	return cl.consumerGroup.Checker()
}

func (cl *Consumer) Close() error {
	return cl.consumerGroup.Close()
}

func (cl *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	cl.logger.WithContext(session.Context()).
		WithField("kafka.generation_id", session.GenerationID()).
		WithField("kafka.member_id", session.MemberID()).
		WithField("claims", session.Claims()).
		Info("ready to consume messages")

	return nil
}

func (cl *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	logger := cl.logger.WithContext(session.Context())
	logger.Debug("clean up consumer claims")

	if cl.cancel != nil {
		cl.cancel()
	}

	return cl.err
}

func (cl *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var ctx context.Context
	ctx, cl.cancel = context.WithCancel(session.Context())

	cl.err = cl.consumeClaimLoop(ctx, session, claim)
	return cl.err
}

func (cl *Consumer) AppendHandler(msgType entities.RequestMessageType, msgHandler messenger.MessageHandler) {
	cl.handler[msgType] = msgHandler
}

func (cl *Consumer) consumeClaimLoop(ctx context.Context, session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger := cl.logger.WithContext(ctx)
	logger.WithField("partition", claim.Partition()).WithField("topic", claim.Topic()).
		Debug("started consuming claims loop")

	for {
		select {
		case <-ctx.Done():
			logger.WithField("reason", ctx.Err().Error()).WithError(cl.err).
				Info("gracefully stopping message claims")
			return cl.err
		case msg, ok := <-claim.Messages():
			// Input channel has been close so we leave the loop
			if !ok {
				return nil
			}

			logger.WithField("timestamp", msg.Timestamp).Trace("message consumed")

			reqMsg, err := NewMessage(msg, session)
			if err != nil {
				errMessage := "failed to decode message request"
				logger.WithError(err).Error(errMessage)
				session.MarkMessage(msg, "")
				continue
			}

			handlerFunc, ok := cl.handler[reqMsg.Type]
			if !ok {
				errMessage := fmt.Sprintf("missing handler for request type %s", reqMsg.Type)
				logger.Error(errMessage)
				session.MarkMessage(msg, "")
				continue
			}

			for _, h := range msg.Headers {
				if string(h.Key) == authutils.UserInfoHeader {
					userInfo := &multitenancy.UserInfo{}
					_ = encoding.Unmarshal(h.Value, userInfo)
					ctx = multitenancy.WithUserInfo(ctx, userInfo)
				}
			}

			err = handlerFunc(ctx, reqMsg)
			if err != nil {
				logger.WithError(err).Error("message has been processed with errors")
				// Invalid req format do not exit loop
				if !errors.IsInvalidFormatError(err) {
					return err
				}
			} else {
				logger.Debug("message has been processed successfully")
			}

			if !cl.disableCommitOnRead {
				session.MarkMessage(msg, "")
				session.Commit()
				cl.logger.WithField("offset", msg.Offset+1).Debug("message has been auto committed")
			}
		}
	}
}
