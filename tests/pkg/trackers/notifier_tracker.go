package trackers

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	infra "github.com/consensys/orchestrate/src/infra/api"
	"github.com/consensys/orchestrate/src/infra/kafka"
	kafkasarama "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	types2 "github.com/consensys/orchestrate/src/notifier/service/types"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
)

const testNotifierTrackerComponent = "tests.notifier.tracker"

type NotifierConsumerTracker struct {
	consumerGroup kafka.ConsumerGroup
	topics        []string
	tracker       *ConsumerTracker
	chanRegistry  *utils.ChanRegistry
	logger        *log.Logger
	cancel        context.CancelFunc
	err           error
}

func NewNotifierConsumerTacker(cfg kafkasarama.Config, topics []string) (*NotifierConsumerTracker, error) {
	chanRegistry := utils.NewChanRegistry()
	cfg.GroupName = testConsumerTrackerComponent + utils.RandString(5)
	consumerGroup, err := kafkasarama.NewConsumerGroup(&cfg)
	if err != nil {
		return nil, err
	}

	logger := log.NewLogger().SetComponent(testNotifierTrackerComponent)
	return &NotifierConsumerTracker{
		consumerGroup: consumerGroup,
		topics:        topics,
		tracker:       NewConsumerTracker(chanRegistry, logger),
		logger:        logger,
		chanRegistry:  chanRegistry,
	}, nil
}

func (cl *NotifierConsumerTracker) Consume(ctx context.Context) error {
	return cl.consumerGroup.Consume(ctx, cl.topics, cl)
}

func (cl *NotifierConsumerTracker) Checker() error {
	return cl.consumerGroup.Checker()
}

func (cl *NotifierConsumerTracker) Close() error {
	return cl.consumerGroup.Close()
}

func (cl *NotifierConsumerTracker) Setup(session sarama.ConsumerGroupSession) error {
	cl.logger.WithContext(session.Context()).
		WithField("kafka.generation_id", session.GenerationID()).
		WithField("kafka.member_id", session.MemberID()).
		WithField("claims", session.Claims()).
		Info("ready to consume messages")

	return nil
}

func (cl *NotifierConsumerTracker) Cleanup(session sarama.ConsumerGroupSession) error {
	logger := cl.logger.WithContext(session.Context())
	logger.Debug("clean up consumer claims")

	if cl.cancel != nil {
		cl.cancel()
	}

	return cl.err
}

func (cl *NotifierConsumerTracker) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var ctx context.Context
	ctx, cl.cancel = context.WithCancel(session.Context())

	cl.err = cl.consumeClaimLoop(ctx, session, claim)
	return cl.err
}

func (m *NotifierConsumerTracker) WaitForMinedTransaction(ctx context.Context, sourceUUID string, timeout time.Duration) (*types2.NotificationResponse, error) {
	msg, err := m.tracker.WaitForMessage(ctx, keyGenOf(sourceUUID, string(entities.NotificationTypeTxMined)), timeout)
	if err != nil {
		return nil, err
	}
	req, ok := msg.(*types2.NotificationResponse)
	if !ok {
		return nil, errors.EncodingError(invalidTypeErr)
	}

	job := &entities.Job{}
	bJob, _ := json.Marshal(req.Data)
	_ = json.Unmarshal(bJob, job)
	req.Data = job
	return req, nil
}

func (m *NotifierConsumerTracker) WaitForFailedTransaction(ctx context.Context, sourceUUID string, timeout time.Duration) (*types2.NotificationResponse, error) {
	msg, err := m.tracker.WaitForMessage(ctx, keyGenOf(sourceUUID, string(entities.NotificationTypeTxFailed)), timeout)
	if err != nil {
		return nil, err
	}
	req, ok := msg.(*types2.NotificationResponse)
	if !ok {
		return nil, errors.EncodingError(invalidTypeErr)
	}
	return req, nil
}

func (cl *NotifierConsumerTracker) consumeClaimLoop(ctx context.Context, session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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

			req := &types2.NotificationResponse{}
			err := infra.UnmarshalBody(bytes.NewReader(msg.Value), req)
			if err != nil {
				errMessage := "failed to decode notification request"
				logger.WithError(err).Error(errMessage)
				session.MarkMessage(msg, "")
				continue
			}

			msgId := keyGenOf(req.SourceUUID, req.Type)
			if !cl.chanRegistry.HasChan(msgId) {
				cl.chanRegistry.Register(msgId, make(chan interface{}, 1))
			}
		
			err = cl.chanRegistry.Send(msgId, req)
			if err != nil {
				logger.WithError(err).Error("message dispatched with errors")
				return err
			}

			logger.Debug("message has been processed successfully")

			session.MarkMessage(msg, "")
			session.Commit()
		}
	}
}
