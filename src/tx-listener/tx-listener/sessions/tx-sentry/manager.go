package txsentry

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/sessions"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const retryJobSessionMngrComponent = "tx-listener.use-case.tx-sentry.session-manager"

type RetryJobSessionMngr struct {
	sendRetryJobUseCase usecases.RetryJob
	messengerAPI        sdk.MessengerAPI
	jobClient           sdk.JobClient
	retrySessionState   store.RetryJobSession
	pendingJobState     store.PendingJob
	logger              *log.Logger
}

func NewRetrySessionManager(messengerAPI sdk.MessengerAPI,
	jobClient sdk.JobClient,
	sendRetryJobUseCase usecases.RetryJob,
	retrySessionState store.RetryJobSession,
	pendingJobState store.PendingJob,
	logger *log.Logger,
) sessions.RetryJobSessionManager {
	return &RetryJobSessionMngr{
		messengerAPI:        messengerAPI,
		sendRetryJobUseCase: sendRetryJobUseCase,
		jobClient:           jobClient,
		retrySessionState:   retrySessionState,
		pendingJobState:     pendingJobState,
		logger:              logger.SetComponent(retryJobSessionMngrComponent),
	}
}

func (mngr *RetryJobSessionMngr) StartSession(ctx context.Context, job *entities.Job) error {
	logger := mngr.logger.
		WithField("job", job.UUID).
		WithField("chain", job.ChainUUID)

	if mngr.retrySessionState.Has(ctx, job.UUID) {
		errMsg := "retry job session already exists"
		logger.Warn(errMsg)
		return errors.AlreadyExistsError(errMsg)
	}

	sess := NewRetryJobSession(mngr.messengerAPI, mngr.jobClient, mngr.sendRetryJobUseCase, mngr.pendingJobState, job, logger)

	err := mngr.retrySessionState.Add(ctx, job)
	if err != nil {
		errMsg := "failed to persist retry session"
		logger.WithError(err).Error(errMsg)
		return err
	}

	go func(sess *RetryJobSession) {
		// @TODO How to propagate error ???
		err := sess.Start(ctx)
		if err != nil {
			errMsg := "failed to run retry session"
			logger.WithError(err).Error(errMsg)
		}

		err = mngr.retrySessionState.Remove(ctx, job.UUID)
		if err != nil {
			errMsg := "failed to remove retry session"
			logger.WithError(err).Error(errMsg)
		}
	}(sess)

	return nil
}
