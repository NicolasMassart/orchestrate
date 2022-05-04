package txsentry

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const retryJobSessionComponent = "tx-listener.retry-job.session"

type RetryJobSession struct {
	sendRetryJobUseCase usecases.RetryJob
	client              sdk.OrchestrateClient
	pendingJobState     store.PendingJob
	logger              *log.Logger
	job                 *entities.Job
	cancelCtx           context.CancelFunc
	cerr                chan error
}

type sessionData struct {
	parentJob        *entities.Job
	nChildren        int
	retries          int
	lastChildJobUUID string
}

func NewRetryJobSession(client sdk.OrchestrateClient, sendRetryJobUseCase usecases.RetryJob, pendingJobState store.PendingJob, job *entities.Job, logger *log.Logger) *RetryJobSession {
	return &RetryJobSession{
		sendRetryJobUseCase: sendRetryJobUseCase,
		client:              client,
		job:                 job,
		pendingJobState:     pendingJobState,
		logger:              logger.WithField("job", job.UUID).SetComponent(retryJobSessionComponent),
		cerr:                make(chan error, 1),
	}
}

func (uc *RetryJobSession) Start(ctx context.Context) error {
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
	}()

	select {
	case err := <-uc.cerr:
		uc.logger.WithField("err", err).Error("exited with errors")
		return err
	case <-ctx.Done():
		uc.logger.WithField("reason", ctx.Err().Error()).Info("gracefully stopping...")
		return nil
	}
}

func (uc *RetryJobSession) Stop() {
	uc.logger.Info("session has been stopped")
	uc.cancelCtx()
}

func (uc *RetryJobSession) runSession(ctx context.Context, sess *sessionData) error {
	ticker := time.NewTicker(sess.parentJob.InternalData.RetryInterval)
	for {
		select {
		case <-ticker.C:
			_, err := uc.pendingJobState.GetByTxHash(ctx, uc.job.ChainUUID, uc.job.Transaction.Hash)
			if err != nil {
				if errors.IsNotFoundError(err) {
					uc.Stop()
					return nil
				}
				return err
			}

			uc.logger.WithField("children", sess.nChildren).WithField("retries", sess.retries).
				Debug("running session iteration")
			childJobUUID, err := uc.sendRetryJobUseCase.Execute(ctx, sess.parentJob, sess.lastChildJobUUID, sess.nChildren)
			if err != nil {
				return err
			}

			sess.retries++
			if sess.retries >= types.SentryMaxRetries {
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

			if childJobUUID != sess.lastChildJobUUID {
				sess.nChildren++
				sess.lastChildJobUUID = childJobUUID
			}
		case <-ctx.Done():
			uc.logger.WithField("reason", ctx.Err().Error()).Info("session gracefully stopped")
			return nil
		}
	}
}

func (uc *RetryJobSession) updateJobAnnotations() error {
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

func (uc *RetryJobSession) retrieveJobSessionData(ctx context.Context, job *entities.Job) (*sessionData, error) {
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

	return &sessionData{
		parentJob:        job,
		nChildren:        nChildren,
		retries:          nRetries,
		lastChildJobUUID: lastJobRetry.UUID,
	}, nil
}
