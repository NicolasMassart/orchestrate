package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/infra/kafka"
	saramainfra "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/messenger"
)

const (
	errorProcessingMessage = "error processing message"
)

type ConsumerMessageHandler interface {
	ID() string
	ProcessMsg(ctx context.Context, rawMsg *sarama.ConsumerMessage, decodedMsg interface{}) error
	DecodeMessage(rawMsg *sarama.ConsumerMessage) (interface{}, error)
}

type Consumer struct {
	consumerGroup kafka.Consumer
	handler       ConsumerMessageHandler
	topics        []string
	cancel        context.CancelFunc
	logger        *log.Logger
	err           error
}

var _ messenger.Consumer = &Consumer{}

func NewMessageConsumer(cfg *saramainfra.Config, topics []string, handler ConsumerMessageHandler) (*Consumer, error) {
	consumerGroup, err := saramainfra.NewConsumerGroup(cfg)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		topics:        topics,
		handler:       handler,
		logger:        log.NewLogger().SetComponent(handler.ID()),
	}, nil
}

func (cl *Consumer) Consume(ctx context.Context) error {
	return cl.consumerGroup.Consume(ctx, cl.topics, cl)
}

func (cl *Consumer) Checker() error {
	return cl.consumerGroup.Client().Checker()
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

func (cl *Consumer) consumeClaimLoop(ctx context.Context, session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger := cl.logger.WithContext(ctx)
	logger.WithField("partition", claim.Partition()).WithField("topic", claim.Topic()).
		Debug("started consuming claims loop")

	for {
		select {
		case <-ctx.Done():
			logger.WithField("reason", ctx.Err().Error()).WithError(cl.err).
				Info("gracefully stopping message cl...")
			return cl.err
		case msg, ok := <-claim.Messages():
			// Input channel has been close so we leave the loop
			if !ok {
				return nil
			}

			logger.WithField("timestamp", msg.Timestamp).Trace("message consumed")

			decodedMsg, err := cl.handler.DecodeMessage(msg)
			if err != nil {
				logger.WithError(err).Error("error decoding message", msg)
				session.MarkMessage(msg, "")
				continue
			}

			err = cl.handler.ProcessMsg(ctx, msg, decodedMsg)
			if err != nil {
				logger.WithError(err).Error(errorProcessingMessage)
				return err
			}

			logger.Debug("message has been processed successfully")
			session.MarkMessage(msg, "")
			session.Commit()
		}
	}
}
