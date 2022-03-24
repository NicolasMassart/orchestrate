package events

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/chain-listener/store"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const chainBlockTxsUseCaseComponent = "chain-listener.use-case.event.chain-block-txs"

type chainBlockTxsUC struct {
	pendingJobState     store.PendingJob
	retrySessionState   store.RetrySessions
	notifyMinedJob      usecases.NotifyMinedJob
	updateChainHead     usecases.UpdateChainHead
	retrySessionManager usecases.RetryJobSessionManager
	logger              *log.Logger
}

func ChainBlockTxsUseCase(notifyMinedJob usecases.NotifyMinedJob,
	updateChainHead usecases.UpdateChainHead,
	retrySessionManager usecases.RetryJobSessionManager,
	pendingJobState store.PendingJob,
	retrySessionState store.RetrySessions,
	logger *log.Logger,
) usecases.ChainBlockTxsUseCase {
	return &chainBlockTxsUC{
		pendingJobState:     pendingJobState,
		retrySessionState:   retrySessionState,
		retrySessionManager: retrySessionManager,
		notifyMinedJob:      notifyMinedJob,
		updateChainHead:     updateChainHead,
		logger:              logger.SetComponent(chainBlockTxsUseCaseComponent),
	}
}

func (uc *chainBlockTxsUC) Execute(ctx context.Context, chainUUID string, blockNumber uint64, txHashes []*ethcommon.Hash) error {
	logger := uc.logger.WithField("block", blockNumber)
	logger.WithField("txs", len(txHashes)).Debug("processing block")

	// @TODO Run in parallel
	for _, txHash := range txHashes {
		if err := uc.handleRetrySessionJob(ctx, chainUUID, txHash); err != nil {
			return err
		}
		if err := uc.handlePendingJob(ctx, chainUUID, txHash); err != nil {
			return err
		}
	}

	// @TODO Can we reduce the amount of updates???
	err := uc.updateChainHead.Execute(ctx, chainUUID, blockNumber)
	if err != nil {
		return err
	}

	return nil
}

func (uc *chainBlockTxsUC) handlePendingJob(ctx context.Context, chainUUID string, txHash *ethcommon.Hash) error {
	logger := uc.logger.WithField("chain", chainUUID).WithField("tx_hash", txHash.String())
	logger.Debug("handling pending job")
	minedJob, err := uc.pendingJobState.GetByTxHash(ctx, chainUUID, txHash)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil
		}
		return err
	}

	err = uc.notifyMinedJob.Execute(ctx, minedJob)
	if err != nil {
		return err
	}
	err = uc.pendingJobState.Remove(ctx, minedJob.UUID)
	if err != nil {
		logger.WithError(err).Error("failed to remove pending job")
		return err
	}

	return nil
}

func (uc *chainBlockTxsUC) handleRetrySessionJob(ctx context.Context, chainUUID string, txHash *ethcommon.Hash) error {
	logger := uc.logger.WithField("chain", chainUUID).WithField("tx_hash", txHash.String())
	logger.Debug("handling retry session job")
	retryJobSessID, err := uc.retrySessionState.GetByTxHash(ctx, chainUUID, txHash)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil
		}
		logger.WithError(err).Error("failed to find retry session")
		return err
	}

	err = uc.retrySessionManager.StopSession(ctx, retryJobSessID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil
		}
		logger.WithError(err).Error("failed to stop retry session")
		return err
	}

	return nil
}
