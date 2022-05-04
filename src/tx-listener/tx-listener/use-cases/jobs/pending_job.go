package jobs

import (
	"context"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const pendingJobUseCaseComponent = "tx-listener.use-case.pending-job"

type pendingJobMsg struct {
	ethClient       ethclient.Client
	pendingJobState store.PendingJob
	apiClient       sdk.OrchestrateClient
	minedJob        usecases.MinedJob
	logger          *log.Logger
}

func PendingJob(apiClient sdk.OrchestrateClient,
	ethClient ethclient.Client,
	minedJob usecases.MinedJob,
	pendingJobState store.PendingJob,
	logger *log.Logger,
) usecases.PendingJob {
	return &pendingJobMsg{
		ethClient:       ethClient,
		pendingJobState: pendingJobState,
		apiClient:       apiClient,
		minedJob:        minedJob,
		logger:          logger.SetComponent(pendingJobUseCaseComponent),
	}
}

func (uc *pendingJobMsg) Execute(ctx context.Context, job *entities.Job) error {
	logger := uc.logger.WithField("job", job.UUID).
		WithField("chain", job.ChainUUID).
		WithField("txHash", job.Transaction.Hash.String()).
		WithField("retry", job.ShouldBeRetried())

	logger.Debug("handling new pending job")

	jobHasBeenUpdated := false
	if curJob, _ := uc.pendingJobState.GetJobUUID(ctx, job.UUID); curJob != nil {
		if curJob.Transaction.Hash.String() == job.Transaction.Hash.String() {
			logger.Warn("skipping already known job")
			return nil
		}
		logger.Warn("duplicated job with different transaction hash")
		jobHasBeenUpdated = true
	}

	proxyURL := uc.apiClient.ChainProxyURL(job.ChainUUID)
	receipt, err := uc.ethClient.TransactionReceipt(ctx, proxyURL, *job.Transaction.Hash)
	if err == nil && receipt != nil {
		err2 := uc.minedJob.Execute(ctx, job)
		if err2 != nil {
			return err2
		}
		return nil
	}

	if jobHasBeenUpdated {
		err = uc.pendingJobState.Update(ctx, job)
		if err != nil {
			logger.WithError(err).Error("failed to update job")
			return err
		}
		logger.Debug("pending job updated successfully")
	} else {
		err = uc.pendingJobState.Add(ctx, job)
		if err != nil {
			logger.WithError(err).Error("failed to persist job")
			return err
		}
		logger.Debug("pending job persisted successfully")
	}

	return nil
}
