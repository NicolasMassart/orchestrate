package events

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/chain-listener/store"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
)

const pendingJobUseCaseComponent = "chain-listener.use-case.event.pending-job"

type pendingJobUC struct {
	ethClient              ethclient.Client
	pendingJobState        store.PendingJob
	apiClient              orchestrateclient.OrchestrateClient
	notifyMinedJob         usecases.NotifyMinedJob
	retryJobSessionManager usecases.RetryJobSessionManager
	logger                 *log.Logger
}

func PendingJobUseCase(apiClient orchestrateclient.OrchestrateClient,
	ethClient ethclient.Client,
	notifyMinedJob usecases.NotifyMinedJob,
	retryJobSessionManager usecases.RetryJobSessionManager,
	pendingJobState store.PendingJob,
	logger *log.Logger,
) usecases.PendingJobUseCase {
	return &pendingJobUC{
		ethClient:              ethClient,
		pendingJobState:        pendingJobState,
		apiClient:              apiClient,
		notifyMinedJob:         notifyMinedJob,
		retryJobSessionManager: retryJobSessionManager,
		logger:                 logger.SetComponent(pendingJobUseCaseComponent),
	}
}

func (uc *pendingJobUC) Execute(ctx context.Context, job *entities.Job) error {
	logger := uc.logger.WithField("job", job.UUID).
		WithField("chain", job.ChainUUID).
		WithField("txHash", job.Transaction.Hash.String()).
		WithField("retry", job.ShouldBeRetried())

	logger.Debug("handling new pending job")

	if curJob, _ := uc.pendingJobState.GetJobUUID(ctx, job.UUID); curJob != nil {
		logger.Warn("skipping already known job")
		return nil
	}

	proxyURL := uc.apiClient.ChainProxyURL(job.ChainUUID)
	receipt, _ := uc.ethClient.TransactionReceipt(ctx, proxyURL, *job.Transaction.Hash)
	if receipt != nil {
		err := uc.notifyMinedJob.Execute(ctx, job)
		if err != nil {
			return err
		}
		return nil
	}

	err := uc.pendingJobState.Add(ctx, job)
	if err != nil {
		logger.WithError(err).Error("failed to persist job")
		return err
	}
	logger.Info("pending job persisted successfully")

	if job.ShouldBeRetried() {
		err := uc.retryJobSessionManager.StartSession(ctx, job)
		if err != nil && !errors.IsAlreadyExistsError(err) {
			return err
		}
	}

	return nil
}
