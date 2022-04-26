package txsentry

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/sessions"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const retryJobSessionMngrComponent = "tx-listener.use-case.tx-sentry.session-manager"

type RetryJobSessionMngr struct {
	sendRetryJobUseCase usecases.RetryJob
	client              orchestrateclient.OrchestrateClient
	retrySessionState   store.RetrySessions
	logger              *log.Logger
	sessions            map[string]*RetryJobSession
	sessionsErr         map[string]error
}

func NewRetrySessionManager(client orchestrateclient.OrchestrateClient,
	sendRetryJobUseCase usecases.RetryJob,
	retrySessionState store.RetrySessions,
	logger *log.Logger,
) sessions.TxSentrySessionManager {
	return &RetryJobSessionMngr{
		sendRetryJobUseCase: sendRetryJobUseCase,
		client:              client,
		retrySessionState:   retrySessionState,
		sessions:            make(map[string]*RetryJobSession),
		sessionsErr:         make(map[string]error),
		logger:              logger.SetComponent(retryJobSessionMngrComponent),
	}
}

func (mngr *RetryJobSessionMngr) StartSession(ctx context.Context, job *entities.Job) error {
	logger := mngr.logger.
		WithField("job", job.UUID).
		WithField("chain", job.ChainUUID)

	if _, ok := mngr.sessions[job.UUID]; ok {
		errMsg := "retry job session already exists"
		logger.Warn(errMsg)
		return errors.AlreadyExistsError(errMsg)
	}

	sess := NewRetryJobSession(mngr.client, mngr.sendRetryJobUseCase, mngr.retrySessionState, job, logger)

	mngr.sessions[job.UUID] = sess
	err := mngr.retrySessionState.Add(ctx, job.UUID, job)
	if err != nil {
		errMsg := "failed to persist retry session"
		logger.WithError(err).Error(errMsg)
		return err
	}

	go func(sess *RetryJobSession, jobUUID string) {
		err := sess.Start(ctx)
		if err != nil {
			mngr.sessionsErr[jobUUID] = err
			errMsg := "failed to run retry session"
			logger.WithError(err).Error(errMsg)
		}
	}(sess, job.UUID)

	return nil
}

func (mngr *RetryJobSessionMngr) StopSession(ctx context.Context, jobUUID string) error {
	logger := mngr.logger.WithField("job", jobUUID)
	if _, ok := mngr.sessions[jobUUID]; !ok {
		errMsg := "retry job session is not found"
		logger.WithField("job", jobUUID).Error(errMsg)
		return errors.NotFoundError(errMsg)
	}

	sess := mngr.sessions[jobUUID]
	if err := sess.Stop(); err != nil {
		return err
	}

	err := mngr.retrySessionState.Remove(ctx, jobUUID)
	if err != nil {
		return err
	}

	delete(mngr.sessions, jobUUID)

	// If an error happened during the Start()
	if err, ok := mngr.sessionsErr[jobUUID]; ok {
		return err
	}

	return nil
}
