package service

import (
	"bytes"
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/api"
	"github.com/consensys/orchestrate/src/tx-listener/service/types"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/sessions"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

var (
	PendingJobMessageType entities.RequestMessageType = "pending-job"
)

type JobHandler struct {
	pendingJobUC        usecases.PendingJob
	failedJobUC         usecases.FailedJob
	retryBackOff        backoff.BackOff
	retryJobSessionMngr sessions.RetryJobSessionManager
	chainSessionMngr    sessions.ChainSessionManager
	completedJobs       *utils.SortedList
	logger              *log.Logger
}

func NewJobHandler(pendingJobUC usecases.PendingJob,
	failedJobUC usecases.FailedJob,
	chainSessionMngr sessions.ChainSessionManager,
	retryJobSessionMngr sessions.RetryJobSessionManager,
	bck backoff.BackOff) *JobHandler {
	return &JobHandler{
		retryJobSessionMngr: retryJobSessionMngr,
		chainSessionMngr:    chainSessionMngr,
		pendingJobUC:        pendingJobUC,
		failedJobUC:         failedJobUC,
		retryBackOff:        bck,
		completedJobs:       utils.NewSortedList(),
		logger:              log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (mch *JobHandler) HandlePendingJobMessage(ctx context.Context, msg *entities.Message) error {
	req := &types.PendingJobMessageRequest{}
	err := api.UnmarshalBody(bytes.NewReader(msg.Body), req)
	if err != nil {
		return errors.InvalidFormatError("invalid pending job request type")
	}

	return backoff.RetryNotify(
		func() error {
			err := mch.processPendingJob(ctx, req, msg)
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
				err = mch.failedJobUC.Execute(ctx, req.Job, err.Error())
				if err != nil {
					return backoff.Permanent(err)
				}
			}

			return nil
		},
		mch.retryBackOff,
		func(err error, duration time.Duration) {
			mch.logger.WithError(err).Warnf("error processing message, retrying in %v...", duration)
		},
	)
}

func (mch *JobHandler) processPendingJob(ctx context.Context, req *types.PendingJobMessageRequest, msg *entities.Message) error {
	logger := mch.logger.WithField("job", req.Job.UUID).WithField("schedule", req.Job.ScheduleUUID)

	// @TODO Use full Chain object to avoid re fetching
	err := mch.chainSessionMngr.StartSession(ctx, req.Job.ChainUUID)
	if err != nil && !errors.IsAlreadyExistsError(err) {
		return err
	}

	err = mch.pendingJobUC.Execute(ctx, req.Job, msg)
	if err != nil {
		logger.WithError(err).Error("failed to handle pending job")
		return err
	}

	if req.Job.ShouldBeRetried() {
		err = mch.retryJobSessionMngr.StartSession(ctx, req.Job)
		if err != nil && !errors.IsAlreadyExistsError(err) {
			logger.WithError(err).Error("failed to start tx-sentry session")
			return err
		}
	}

	return nil
}
