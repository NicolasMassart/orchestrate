package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-sender/service/types"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/utils"

	"github.com/consensys/orchestrate/pkg/errors"
	usecases "github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases"
)

var StartedJobMessageType entities.RequestMessageType = "started-job"

type JobHandler struct {
	useCases     usecases.UseCases
	retryBackOff backoff.BackOff
	messengerAPI sdk.MessengerAPI
	logger       *log.Logger
}

func NewJobHandler(useCases usecases.UseCases, messengerAPI sdk.MessengerAPI, bck backoff.BackOff) *JobHandler {
	return &JobHandler{
		useCases:     useCases,
		retryBackOff: bck,
		messengerAPI: messengerAPI,
		logger:       log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (mch *JobHandler) HandleStartedJob(ctx context.Context, msg *entities.Message) error {
	req := &types.StartedJobReq{}
	err := json.Unmarshal(msg.Body, req)
	if err != nil {
		mch.logger.Warnf("%q", msg.Body)
		return errors.InvalidFormatError("invalid start job request type")
	}

	logger := mch.logger.WithField("job", req.Job.UUID).WithField("schedule", req.Job.ScheduleUUID)
	err = backoff.RetryNotify(
		func() error {
			err = mch.executeSendJob(ctx, req.Job)
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
			case req.Job.InternalData.ParentJobUUID != "":
				if req.Job.InternalData.ParentJobUUID == req.Job.UUID {
					logger.WithError(err).Warn("ignoring errors on resending jobs...")
					return nil
				}
				serr = utils.UpdateJobStatus(ctx, mch.messengerAPI, req.Job,
					entities.StatusFailed, err.Error(), nil)
			// Retry over same message
			case errors.IsKnownTransactionError(err) || errors.IsNonceTooLowError(err):
				return err
			case errors.IsInvalidNonceWarning(err):
				resetJobTx(req.Job)
				serr = utils.UpdateJobStatus(ctx, mch.messengerAPI, req.Job,
					entities.StatusRecovering, err.Error(), nil)
			default:
				serr = utils.UpdateJobStatus(ctx, mch.messengerAPI, req.Job,
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

	if err != nil {
		serr := utils.UpdateJobStatus(ctx, mch.messengerAPI, req.Job, entities.StatusFailed, err.Error(), nil)
		if serr != nil {
			return serr
		}
	}

	return nil
}

func (mch *JobHandler) executeSendJob(ctx context.Context, job *entities.Job) error {
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
