package service

import (
	"context"
	encoding "encoding/json"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/errors"
	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	saramainfra "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	messenger "github.com/consensys/orchestrate/src/infra/messenger/kafka"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/sessions"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const (
	messageListenerComponent = "service.kafka-consumer"
)

func NewMessageConsumer(cfg *saramainfra.Config,
	topics []string,
	pendingJobUC usecases.PendingJob,
	failedJobUC usecases.FailedJob,
	chainSessionMngr sessions.ChainSessionManager,
	retryJobSessionMngr sessions.RetryJobSessionManager,
	bck backoff.BackOff,
) (*messenger.Consumer, error) {
	msgConsumerHandler := newMessageConsumerHandler(pendingJobUC, failedJobUC, chainSessionMngr, retryJobSessionMngr, bck)
	return messenger.NewMessageConsumer(cfg, topics, msgConsumerHandler)
}

type messageConsumerHandler struct {
	pendingJobUC        usecases.PendingJob
	failedJobUC         usecases.FailedJob
	retryBackOff        backoff.BackOff
	retryJobSessionMngr sessions.RetryJobSessionManager
	chainSessionMngr    sessions.ChainSessionManager
	logger              *log.Logger
}

var _ messenger.ConsumerMessageHandler = &messageConsumerHandler{}

func newMessageConsumerHandler(pendingJobUC usecases.PendingJob,
	failedJobUC usecases.FailedJob,
	chainSessionMngr sessions.ChainSessionManager,
	retryJobSessionMngr sessions.RetryJobSessionManager,
	bck backoff.BackOff) *messageConsumerHandler {
	return &messageConsumerHandler{
		retryJobSessionMngr: retryJobSessionMngr,
		chainSessionMngr:    chainSessionMngr,
		pendingJobUC:        pendingJobUC,
		failedJobUC:         failedJobUC,
		retryBackOff:        bck,
		logger:              log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (mch *messageConsumerHandler) ProcessMsg(ctx context.Context, rawMsg *sarama.ConsumerMessage, decodedMsg interface{}) error {
	job := decodedMsg.(*entities.Job)
	logger := mch.logger.WithField("job", job.UUID).WithField("schedule", job.ScheduleUUID)
	for _, h := range rawMsg.Headers {
		if string(h.Key) == authutils.UserInfoHeader {
			userInfo := &multitenancy.UserInfo{}
			_ = encoding.Unmarshal(h.Value, userInfo)
			ctx = multitenancy.WithUserInfo(ctx, userInfo)
		}
	}

	return backoff.RetryNotify(
		func() error {
			err := mch.processTask(ctx, job, logger)
			switch {
			// Exits if not errors
			case err == nil:
				return nil
			case err == context.DeadlineExceeded || err == context.Canceled:
				return backoff.Permanent(ctx.Err())
			case ctx.Err() != nil:
				return backoff.Permanent(ctx.Err())
			case errors.IsConnectionError(err):
				return err
			default: // Remaining error types (err != nil)
				err = mch.failedJobUC.Execute(ctx, job, err.Error())
				if err != nil {
					return backoff.Permanent(err)
				}
				return nil
			}
		},
		mch.retryBackOff,
		func(err error, duration time.Duration) {
			logger.WithError(err).Warnf("error processing job, retrying in %v...", duration)
		},
	)
}

func (mch *messageConsumerHandler) DecodeMessage(rawMsg *sarama.ConsumerMessage) (interface{}, error) {
	return messenger.DecodeJobMessage(rawMsg)
}

func (mch *messageConsumerHandler) ID() string {
	return messageListenerComponent
}

func (mch *messageConsumerHandler) processTask(ctx context.Context, job *entities.Job, logger *log.Logger) error {
	err := mch.pendingJobUC.Execute(ctx, job)
	if err != nil {
		logger.WithError(err).Error("failed to handle pending job")
		return err
	}

	if job.ShouldBeRetried() {
		err = mch.retryJobSessionMngr.StartSession(ctx, job)
		if err != nil && !errors.IsAlreadyExistsError(err) {
			logger.WithError(err).Error("failed to start tx-sentry session")
			return err
		}
	}

	err = mch.chainSessionMngr.StartSession(ctx, job.ChainUUID)
	if err != nil && !errors.IsAlreadyExistsError(err) {
		logger.WithError(err).Error("failed to start chain mch session")
		return err
	}

	return nil
}
