package service

import (
	"context"
	encoding "encoding/json"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	saramainfra "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	messenger "github.com/consensys/orchestrate/src/infra/messenger/kafka"
	utils2 "github.com/consensys/orchestrate/src/tx-sender/tx-sender/utils"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	usecases "github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases"
)

const (
	messageListenerComponent = "service.kafka-consumer"
)

func NewMessageConsumer(cfg *saramainfra.Config, topics []string, useCases usecases.UseCases, jobClient client.JobClient, bck backoff.BackOff) (*messenger.Consumer, error) {
	msgConsumerHandler := newMessageConsumerHandler(useCases, jobClient, bck)
	return messenger.NewMessageConsumer(cfg, topics, msgConsumerHandler)
}

type messageConsumerHandler struct {
	useCases     usecases.UseCases
	retryBackOff backoff.BackOff
	jobClient    client.JobClient
	logger       *log.Logger
}

func newMessageConsumerHandler(useCases usecases.UseCases, jobClient client.JobClient, bck backoff.BackOff) *messageConsumerHandler {
	return &messageConsumerHandler{
		useCases:     useCases,
		retryBackOff: bck,
		jobClient:    jobClient,
		logger:       log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (mch *messageConsumerHandler) DecodeMessage(rawMsg *sarama.ConsumerMessage) (interface{}, error) {
	return messenger.DecodeJobMessage(rawMsg)
}

func (mch *messageConsumerHandler) ID() string {
	return messageListenerComponent
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

	err := mch.processTask(ctx, job, logger)
	if err != nil {
		serr := utils2.UpdateJobStatus(ctx, mch.jobClient, job, entities.StatusFailed, err.Error(), nil)
		if serr != nil {
			return serr
		}
	}

	return nil
}

func (mch *messageConsumerHandler) processTask(ctx context.Context, job *entities.Job, logger *log.Logger) error {
	return backoff.RetryNotify(
		func() error {
			err := mch.executeSendJob(ctx, job)
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
			}

			var serr error
			switch {
			// Retry over same message
			case errors.IsInvalidNonceWarning(err):
				resetJobTx(job)
				serr = utils2.UpdateJobStatus(ctx, mch.jobClient, job,
					entities.StatusRecovering, err.Error(), nil)
				if serr == nil {
					return err
				}
			case errors.IsKnownTransactionError(err) || errors.IsNonceTooLowError(err):
				if job.InternalData.ParentJobUUID != "" {
					logger.WithError(err).Warn("ignoring known transaction or nonce too low when it is a child job...")
					return nil
				}
				return err
			default:
				serr = utils2.UpdateJobStatus(ctx, mch.jobClient, job,
					entities.StatusFailed, err.Error(), nil)
			}

			switch {
			case serr != nil && ctx.Err() != nil: // If context has been cancel, exits
				return backoff.Permanent(ctx.Err())
			case serr != nil && errors.IsConnectionError(serr): // Retry on connection error
				return serr
			case serr != nil: // Other kind of error, we exit
				return backoff.Permanent(serr)
			default:
				return nil
			}
		},
		mch.retryBackOff,
		func(err error, duration time.Duration) {
			logger.WithError(err).Warnf("error processing job, retrying in %v...", duration)
		},
	)
}

func (mch *messageConsumerHandler) executeSendJob(ctx context.Context, job *entities.Job) error {
	switch job.Type {
	case entities.GoQuorumPrivateTransaction:
		return mch.useCases.SendGoQuorumPrivateTx().Execute(ctx, job)
	case entities.GoQuorumMarkingTransaction:
		return mch.useCases.SendGoQuorumMarkingTx().Execute(ctx, job)
	case entities.EEAPrivateTransaction:
		return mch.useCases.SendEEAPrivateTx().Execute(ctx, job)
	case entities.EthereumRawTransaction:
		return mch.useCases.SendETHRawTx().Execute(ctx, job)
	case entities.EEAMarkingTransaction, entities.EthereumTransaction:
		return mch.useCases.SendETHTx().Execute(ctx, job)
	default:
		return errors.InvalidParameterError("job type %s is not supported", job.Type)
	}
}

func resetJobTx(job *entities.Job) {
	job.Transaction.Nonce = nil
	job.Transaction.Hash = nil
	job.Transaction.Raw = nil
}
