package builder

import (
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/jobs"
)

type jobUCs struct {
	pendingJob usecases.PendingJob
	minedJob   usecases.MinedJob
	retryJob   usecases.RetryJob
}

func (b *jobUCs) PendingJobUseCase() usecases.PendingJob {
	return b.pendingJob
}

func (b *jobUCs) MinedJobUseCase() usecases.MinedJob {
	return b.minedJob
}

func (b *jobUCs) RetryJobUseCase() usecases.RetryJob {
	return b.retryJob
}

func NewJobUseCases(apiClient orchestrateclient.OrchestrateClient,
	ethClient ethclient.MultiClient,
	contractUCs usecases.ContractsUseCases,
	state store.State,
	logger *log.Logger,
) usecases.JobUseCases {
	minedJobUC := jobs.MinedJobUseCase(apiClient, ethClient, contractUCs.RegisterDeployedContractUseCase(), logger)
	pendingJobUC := jobs.PendingJob(apiClient, ethClient, minedJobUC, state.PendingJobState(), logger)
	retryJobUC := jobs.RetrySessionJobUseCase(apiClient, logger)

	return &jobUCs{
		pendingJob: pendingJobUC,
		minedJob: minedJobUC,
		retryJob: retryJobUC,
	}
}
