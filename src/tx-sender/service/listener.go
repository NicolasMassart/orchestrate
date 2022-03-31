package service

import (
	"context"
	encoding "encoding/json"
	"time"

	client2 "github.com/consensys/quorum-key-manager/pkg/client"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	utils2 "github.com/consensys/orchestrate/src/tx-sender/tx-sender/utils"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	usecases "github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases"
)

const (
	messageListenerComponent = "service.kafka-consumer"
	errorProcessingMessage   = "error processing message"
)

type MessageListener struct {
	useCases     usecases.UseCases
	retryBackOff backoff.BackOff
	jobClient    client.JobClient
	cancel       context.CancelFunc
	err          error
	logger       *log.Logger
}

func NewMessageListener(useCases usecases.UseCases, jobClient client.JobClient, bck backoff.BackOff) *MessageListener {
	return &MessageListener{
		useCases:     useCases,
		retryBackOff: bck,
		jobClient:    jobClient,
		logger:       log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (listener *MessageListener) Setup(session sarama.ConsumerGroupSession) error {
	listener.logger.WithContext(session.Context()).
		WithField("kafka.generation_id", session.GenerationID()).
		WithField("kafka.member_id", session.MemberID()).
		WithField("claims", session.Claims()).
		Info("ready to consume messages")

	return nil
}

func (listener *MessageListener) Cleanup(session sarama.ConsumerGroupSession) error {
	logger := listener.logger.WithContext(session.Context())
	logger.Info("all claims consumed")
	if listener.cancel != nil {
		logger.Debug("canceling context")
		listener.cancel()
	}

	return listener.err
}

func (listener *MessageListener) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var ctx context.Context
	ctx, listener.cancel = context.WithCancel(session.Context())
	listener.err = listener.consumeClaimLoop(ctx, session, claim)
	return listener.err
}

func (listener *MessageListener) consumeClaimLoop(ctx context.Context, session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger := listener.logger.WithContext(ctx)
	logger.Info("started consuming claims loop")

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

			job, err := decodeMessage(msg)
			if err != nil {
				logger.WithError(err).Error("error decoding message", msg)
				session.MarkMessage(msg, "")
				continue
			}

			jlogger := logger.WithField("job", job.UUID).WithField("schedule", job.ScheduleUUID)
			jlogger.WithField("timestamp", msg.Timestamp).Debug("message consumed")

			ctx = log.With(ctx, jlogger)
			for _, h := range msg.Headers {
				if string(h.Key) == authutils.AuthorizationHeader {
					ctx = appendAuthHeader(ctx, string(h.Value))
				}
			}

			err = listener.processTask(ctx, job)

			if err != nil {
				serr := utils2.UpdateJobStatus(ctx, listener.jobClient, job, entities.StatusFailed, err.Error(), nil)
				if serr != nil {
					jlogger.WithError(serr).Error(errorProcessingMessage)
					return serr
				}
			}

			jlogger.Debug("job message has been processed")
			session.MarkMessage(msg, "")
			session.Commit()
		}
	}
}

func (listener *MessageListener) processTask(ctx context.Context, job *entities.Job) error {
	logger := log.FromContext(ctx)
	return backoff.RetryNotify(
		func() error {
			err := listener.executeSendJob(ctx, job)
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
				serr = utils2.UpdateJobStatus(ctx, listener.jobClient, job,
					entities.StatusRecovering, err.Error(), nil)
				if serr == nil {
					return err
				}
			case errors.IsKnownTransactionError(err):
				// Ignore
				return nil
			default:
				serr = utils2.UpdateJobStatus(ctx, listener.jobClient, job,
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
		listener.retryBackOff,
		func(err error, duration time.Duration) {
			logger.WithError(err).Warnf("error processing job, retrying in %v...", duration)
		},
	)
}

func (listener *MessageListener) executeSendJob(ctx context.Context, job *entities.Job) error {
	switch job.Type {
	case entities.GoQuorumPrivateTransaction:
		return listener.useCases.SendGoQuorumPrivateTx().Execute(ctx, job)
	case entities.GoQuorumMarkingTransaction:
		return listener.useCases.SendGoQuorumMarkingTx().Execute(ctx, job)
	case entities.EEAPrivateTransaction:
		return listener.useCases.SendEEAPrivateTx().Execute(ctx, job)
	case entities.EthereumRawTransaction:
		return listener.useCases.SendETHRawTx().Execute(ctx, job)
	case entities.EEAMarkingTransaction, entities.EthereumTransaction:
		return listener.useCases.SendETHTx().Execute(ctx, job)
	default:
		return errors.InvalidParameterError("job type %s is not supported", job.Type)
	}
}

func decodeMessage(msg *sarama.ConsumerMessage) (*entities.Job, error) {
	job := &entities.Job{}
	err := encoding.Unmarshal(msg.Value, job)
	if err != nil {
		errMessage := "failed to decode job message"
		return nil, errors.EncodingError(errMessage).ExtendComponent(messageListenerComponent)
	}
	return job, nil
}

func resetJobTx(job *entities.Job) {
	job.Transaction.Nonce = nil
	job.Transaction.Hash = nil
	job.Transaction.Raw = nil
}

func appendAuthHeader(ctx context.Context, authHeader string) context.Context {
	return context.WithValue(ctx, client2.RequestHeaderKey, map[string]string{
		authutils.AuthorizationHeader: authHeader,
	})
}
