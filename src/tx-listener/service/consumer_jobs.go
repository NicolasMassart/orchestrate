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
	messenger "github.com/consensys/orchestrate/src/infra/messenger/kafka"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/sessions"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const (
	messageListenerComponent = "service.kafka-consumer"
)

type MessageConsumerHandler struct {
	pendingJobUC     usecases.PendingJob
	retryBackOff     backoff.BackOff
	retrySessionMngr sessions.TxSentrySessionManager
	chainMngr        sessions.ChainSessionManager
	logger           *log.Logger
}

var _ messenger.ConsumerMessageHandler = &MessageConsumerHandler{}

func NewMessageConsumerHandler(pendingJobUC usecases.PendingJob,
	chainMngr sessions.ChainSessionManager,
	retrySessionMngr sessions.TxSentrySessionManager,
	bck backoff.BackOff) *MessageConsumerHandler {
	return &MessageConsumerHandler{
		retrySessionMngr: retrySessionMngr,
		chainMngr:        chainMngr,
		pendingJobUC:     pendingJobUC,
		retryBackOff:     bck,
		logger:           log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (mch *MessageConsumerHandler) ProcessMsg(ctx context.Context, rawMsg *sarama.ConsumerMessage, decodedMsg interface{}) error {
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
				return backoff.Permanent(err)
			case ctx.Err() != nil:
				return backoff.Permanent(ctx.Err())
			case errors.IsConnectionError(err):
				return err
			default:
				return backoff.Permanent(err)
			}
		},
		mch.retryBackOff,
		func(err error, duration time.Duration) {
			logger.WithError(err).Warnf("error processing job, retrying in %v...", duration)
		},
	)
}

func (mch *MessageConsumerHandler) DecodeMessage(rawMsg *sarama.ConsumerMessage) (interface{}, error) {
	return messenger.DecodeJobMessage(rawMsg)
}

func (mch *MessageConsumerHandler) ID() string {
	return messageListenerComponent
}

func (mch *MessageConsumerHandler) processTask(ctx context.Context, job *entities.Job, logger *log.Logger) error {
	err := mch.pendingJobUC.Execute(ctx, job)
	if err != nil {
		logger.WithError(err).Error("failed to handle pending job")
		return err
	}

	if job.ShouldBeRetried() {
		err2 := mch.retrySessionMngr.StartSession(ctx, job)
		if err2 != nil && !errors.IsAlreadyExistsError(err2) {
			logger.WithError(err2).Error("failed to start tx-sentry session")
			return err2
		}
	}

	err = mch.chainMngr.StartSession(ctx, job.ChainUUID)
	if err != nil {
		logger.WithError(err).Error("failed to start chain mch session")
		return err
	}

	return nil
}
