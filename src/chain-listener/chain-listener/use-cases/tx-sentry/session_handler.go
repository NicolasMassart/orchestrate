package txsentry

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	"github.com/consensys/orchestrate/src/api/service/types"
	usecases "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/chain-listener/store"
	"github.com/consensys/orchestrate/src/entities"
)

const retryJobSessionComponent = "chain-listener.use-case.tx-sentry.session.handler"

// retryJobSession is a manager of job sessions
type retryJobSessionMngr struct {
	sendRetryJobUseCase usecases.SendRetryJob
	client              orchestrateclient.OrchestrateClient
	retrySessionState   store.RetrySessions
	logger              *log.Logger
	sessions            map[string]*retryJobSession
	sessionsErr         map[string]error
}

func RetrySessionManager(client orchestrateclient.OrchestrateClient,
	sendRetryJobUseCase usecases.SendRetryJob,
	retrySessionState store.RetrySessions,
	logger *log.Logger,
) usecases.RetryJobSessionManager {
	return &retryJobSessionMngr{
		sendRetryJobUseCase: sendRetryJobUseCase,
		client:              client,
		retrySessionState:   retrySessionState,
		sessions:            make(map[string]*retryJobSession),
		sessionsErr:         make(map[string]error),
		logger:              logger.SetComponent(retryJobSessionComponent),
	}
}

func (uc *retryJobSessionMngr) StartSession(ctx context.Context, job *entities.Job) error {
	logger := uc.logger.
		WithField("job", job.UUID).
		WithField("chain", job.ChainUUID)

	if _, ok := uc.sessions[job.UUID]; ok {
		errMsg := "session already exists"
		logger.Warn(errMsg)
		return errors.AlreadyExistsError(errMsg)
	}

	sess := &retryJobSession{
		sendRetryJobUseCase: uc.sendRetryJobUseCase,
		client:              uc.client,
		job:                 job,
		logger:              logger,
		cerr:                make(chan error, 1),
	}

	uc.sessions[job.UUID] = sess
	err := uc.retrySessionState.Add(ctx, job.UUID, job)
	if err != nil {
		errMsg := "failed to persist retry session"
		logger.WithError(err).Error(errMsg)
		return err
	}

	go func(sess *retryJobSession, jobUUID string) {
		err := sess.Start(ctx)
		if err != nil {
			uc.sessionsErr[jobUUID] = err
			errMsg := "failed to run retry session"
			logger.WithError(err).Error(errMsg)
		}
	}(sess, job.UUID)

	return nil
}

func (uc *retryJobSessionMngr) StopChainSessions(ctx context.Context, chainUUID string) error {
	logger := uc.logger.WithField("chain", chainUUID)
	jobUUIDs, err := uc.retrySessionState.ListByChainUUID(ctx, chainUUID)
	if err != nil && !errors.IsNotFoundError(err) {
		logger.WithError(err).Error("failed to retrieve all active retry job sessions per chain")
		return err
	}
	for _, jobUUID := range jobUUIDs {
		if err2 := uc.StopSession(ctx, jobUUID); err2 != nil {
			return err2
		}
	}

	return nil
}

func (uc *retryJobSessionMngr) StopSession(ctx context.Context, jobUUID string) error {
	logger := uc.logger.WithField("job", jobUUID)
	if _, ok := uc.sessions[jobUUID]; !ok {
		errMsg := "retry job session is not found"
		logger.WithField("job", jobUUID).Error(errMsg)
		return errors.NotFoundError(errMsg)
	}

	sess := uc.sessions[jobUUID]
	if err := sess.Stop(); err != nil {
		return err
	}

	err := uc.retrySessionState.Remove(ctx, jobUUID)
	if err != nil {
		return err
	}

	delete(uc.sessions, jobUUID)

	// If an error happened during the Start()
	if err, ok := uc.sessionsErr[jobUUID]; ok {
		return err
	}

	return nil
}

type retryJobSession struct {
	sendRetryJobUseCase usecases.SendRetryJob
	client              orchestrateclient.OrchestrateClient
	logger              *log.Logger
	job                 *entities.Job
	cancelCtx           context.CancelFunc
	cerr                chan error
}

type retryJobSessionData struct {
	parentJob        *entities.Job
	nChildren        int
	retries          int
	lastChildJobUUID string
}

func (uc *retryJobSession) Start(ctx context.Context) error {
	uc.logger.Info("retry session started")
	ctx, uc.cancelCtx = context.WithCancel(ctx)

	ses, err := uc.retrieveJobSessionData(ctx, uc.job)
	if err != nil {
		uc.logger.WithError(err).Error("job listening session failed to start")
		return err
	}

	if ses.retries >= types.SentryMaxRetries {
		uc.logger.Warn("job already reached max retries")
		return nil
	}

	go func() {
		err := uc.runSession(ctx, ses)
		if err != nil {
			uc.cerr <- err
			return
		}

		_ = uc.Stop()
	}()

	select {
	case err := <-uc.cerr:
		if errors.IsInvalidStateError(err) {
			uc.logger.WithField("err", err).Warn("exited with warning")
			return nil
		}
		uc.logger.WithField("err", err).Error("exited with errors")
		return err
	case <-ctx.Done():
		uc.logger.WithField("reason", ctx.Err().Error()).Info("gracefully stopping...")
		return nil
	}
}

func (uc *retryJobSession) Stop() error {
	uc.logger.Info("session has been stopped")
	uc.cancelCtx()
	return nil
}

func (uc *retryJobSession) runSession(ctx context.Context, ses *retryJobSessionData) error {
	ticker := time.NewTicker(ses.parentJob.InternalData.RetryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			uc.logger.
				WithField("children", ses.nChildren).
				WithField("retries", ses.retries).
				Debug("running session iteration")
			childJobUUID, err := uc.sendRetryJobUseCase.Execute(ctx, ses.parentJob, ses.lastChildJobUUID, ses.nChildren)
			if err != nil {
				return err
			}

			ses.retries++
			if ses.retries >= types.SentryMaxRetries {
				err = uc.updateJobAnnotations()
				if err != nil {
					return err
				}
				return nil
			}

			// If no child created but no error, we exit the session gracefully
			if childJobUUID == "" {
				return nil
			}

			if childJobUUID != ses.lastChildJobUUID {
				ses.nChildren++
				ses.lastChildJobUUID = childJobUUID
			}
		case <-ctx.Done():
			uc.logger.WithField("reason", ctx.Err().Error()).Info("session gracefully stopped")
			return nil
		}
	}
}

func (uc *retryJobSession) updateJobAnnotations() error {
	annotations := formatters.FormatInternalDataToAnnotations(uc.job.InternalData)
	annotations.HasBeenRetried = true
	_, err := uc.client.UpdateJob(context.Background(), uc.job.UUID, &types.UpdateJobRequest{
		Annotations: &annotations,
	})
	if err != nil {
		uc.logger.WithError(err).Error("failed to update job labels")
		return err
	}

	uc.logger.Info("job labels has been updated")
	return nil
}

func (uc *retryJobSession) retrieveJobSessionData(ctx context.Context, job *entities.Job) (*retryJobSessionData, error) {
	childrenJobs, err := uc.client.SearchJob(ctx, &entities.JobFilters{
		ChainUUID:     job.ChainUUID,
		ParentJobUUID: job.UUID,
		WithLogs:      true,
	})

	if err != nil {
		return nil, err
	}

	var nChildren int
	var lastJobRetry *entities.Job
	if len(childrenJobs) == 0 {
		nChildren = 0
		lastJobRetry = job
	} else {
		nChildren = len(childrenJobs) - 1
		lastJobRetry = formatters.JobResponseToEntity(childrenJobs[nChildren])
	}

	// we count the number of resending of last job as retries
	nRetries := nChildren
	for _, lg := range lastJobRetry.Logs {
		if lg.Status == entities.StatusResending {
			nRetries++
		}
	}

	return &retryJobSessionData{
		parentJob:        job,
		nChildren:        nChildren,
		retries:          nRetries,
		lastChildJobUUID: lastJobRetry.UUID,
	}, nil
}
