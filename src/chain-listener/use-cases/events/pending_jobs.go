package events

import (
	"context"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/chain-listener/state"
	usecases "github.com/consensys/orchestrate/src/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
)

const pendingJobEventUseCaseComponent = "chain-listener.use-case.event.pending-job"

type pendingJobEvent struct {
	ethClient       ethclient.Client
	pendingJobState state.PendingJob
	apiClient       orchestrateclient.OrchestrateClient
	updateJobStatus usecases.UpdatedJobStatus
	sessionHandler  usecases.SessionHandler
	logger          *log.Logger
}

func PendingJobEventHandler(apiClient orchestrateclient.OrchestrateClient,
	ethClient ethclient.Client,
	updateJobStatus usecases.UpdatedJobStatus,
	sessionHandler usecases.SessionHandler,
	pendingJobState state.PendingJob,
	logger *log.Logger,
) usecases.PendingJobEventHandler {
	return &pendingJobEvent{
		ethClient:       ethClient,
		pendingJobState: pendingJobState,
		apiClient:       apiClient,
		updateJobStatus: updateJobStatus,
		sessionHandler:  sessionHandler,
		logger:          logger.SetComponent(pendingJobEventUseCaseComponent),
	}
}

func (uc *pendingJobEvent) Execute(ctx context.Context, job *entities.Job) error {
	logger := uc.logger.WithField("job", job.UUID).
		WithField("chain", job.ChainUUID).
		WithField("txHash", job.Transaction.Hash.String())

	shouldRetryJob := uc.shouldRetryJob(job)

	logger.WithField("retry", shouldRetryJob).Debug("handling new pending job")

	proxyURL := uc.apiClient.ChainProxyURL(job.ChainUUID)
	receipt, _ := uc.ethClient.TransactionReceipt(ctx, proxyURL, *job.Transaction.Hash)
	if receipt != nil {
		job.Receipt = receipt
		err := uc.updateJobStatus.Execute(ctx, job)
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

	if shouldRetryJob {
		err := uc.sessionHandler.StartSession(ctx, job)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uc *pendingJobEvent) shouldRetryJob(job *entities.Job) bool {
	if job.InternalData.ParentJobUUID != "" {
		return false
	}

	if job.InternalData.RetryInterval == 0 {
		return false
	}

	if job.InternalData.HasBeenRetried {
		return false
	}

	return true
}
