package builder

import (
	"github.com/Shopify/sarama"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases/events"
	tx_listener "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases/tx-listener"
	tx_sentry "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases/tx-sentry"
	in_memory "github.com/consensys/orchestrate/src/chain-listener/store/in-memory"
	"github.com/consensys/orchestrate/src/infra/ethclient"
)

type Builder struct {
	pendingJobUseCase    usecases.PendingJobUseCase
	chainBlockTxsUseCase usecases.ChainBlockTxsUseCase
	addChainUseCase      usecases.AddChainUseCase
	updateChainUseCase   usecases.UpdateChainUseCase
	deleteChainUseCase   usecases.DeleteChainUseCase
}

func (b *Builder) PendingJobUseCase() usecases.PendingJobUseCase {
	return b.pendingJobUseCase
}
func (b *Builder) ChainBlockTxsUseCase() usecases.ChainBlockTxsUseCase {
	return b.chainBlockTxsUseCase
}
func (b *Builder) AddChainUseCase() usecases.AddChainUseCase {
	return b.addChainUseCase
}
func (b *Builder) UpdateChainUseCase() usecases.UpdateChainUseCase {
	return b.updateChainUseCase
}
func (b *Builder) DeleteChainUseCase() usecases.DeleteChainUseCase {
	return b.deleteChainUseCase
}

func NewEventUseCases(apiClient orchestrateclient.OrchestrateClient,
	saramaCli sarama.SyncProducer,
	ethClient ethclient.MultiClient,
	txDecodedTopic string,
	logger *log.Logger,
) *Builder {
	chainInMemory := in_memory.NewChainInMemory()
	pendingJobInMemory := in_memory.NewPendingJobInMemory()
	retrySessionInMemory := in_memory.NewRetrySessionInMemory()

	sendNotificationUC := tx_listener.SendNotificationUseCase(apiClient, saramaCli,
		txDecodedTopic, logger)
	registerDeployedContractUC := tx_listener.RegisterDeployedContractUseCase(apiClient, ethClient, chainInMemory, logger)
	updateJobStatusUC := tx_listener.NotifyMinedJobUseCase(apiClient, ethClient, sendNotificationUC,
		registerDeployedContractUC, chainInMemory, logger)
	updateChainHeadUC := tx_listener.UpdateChainHeadUseCase(apiClient, logger)
	retrySessionJobUC := tx_sentry.RetrySessionJobUseCase(apiClient, logger)
	retryJobSessionManager := tx_sentry.RetrySessionManager(apiClient, retrySessionJobUC, retrySessionInMemory, logger)

	addChainEventUC := events.AddChainUseCase(chainInMemory, logger)
	updateChainEventUC := events.UpdateChainUseCase(chainInMemory, logger)
	deleteChainEventUC := events.DeleteChainUseCase(retryJobSessionManager, chainInMemory, pendingJobInMemory, logger)
	chainBlockEventUC := events.ChainBlockTxsUseCase(updateJobStatusUC, updateChainHeadUC, retryJobSessionManager,
		pendingJobInMemory, retrySessionInMemory, logger)
	pendingJobEventUC := events.PendingJobUseCase(apiClient, ethClient, updateJobStatusUC, retryJobSessionManager,
		pendingJobInMemory, logger)

	return &Builder{
		pendingJobUseCase: pendingJobEventUC,
		chainBlockTxsUseCase: chainBlockEventUC,
		addChainUseCase: addChainEventUC,
		updateChainUseCase: updateChainEventUC,
		deleteChainUseCase: deleteChainEventUC,
	}
}
