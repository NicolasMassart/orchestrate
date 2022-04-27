package chains

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const chainBlockTxsUseCaseComponent = "tx-listener.use-case.new-chain-block"

type chainBlockTxsUC struct {
	pendingJobState store.PendingJob
	minedJob        usecases.MinedJob
	logger          *log.Logger
}

func NewChainBlockUseCase(notifyMinedJob usecases.MinedJob,
	pendingJobState store.PendingJob,
	logger *log.Logger,
) usecases.ChainBlock {
	return &chainBlockTxsUC{
		pendingJobState: pendingJobState,
		minedJob:        notifyMinedJob,
		logger:          logger.SetComponent(chainBlockTxsUseCaseComponent),
	}
}

func (uc *chainBlockTxsUC) Execute(ctx context.Context, chainUUID string, blockNumber uint64, txHashes []*ethcommon.Hash) error {
	logger := uc.logger.WithField("block", blockNumber)
	logger.WithField("txs", len(txHashes)).Debug("processing block")

	// @TODO Run in parallel
	for _, txHash := range txHashes {
		if err := uc.handlePendingJob(ctx, chainUUID, txHash); err != nil {
			return err
		}
	}

	return nil
}

func (uc *chainBlockTxsUC) handlePendingJob(ctx context.Context, chainUUID string, txHash *ethcommon.Hash) error {
	logger := uc.logger.WithField("chain", chainUUID).WithField("tx_hash", txHash.String())
	logger.Debug("handling pending job")
	minedJob, err := uc.pendingJobState.GetByTxHash(ctx, chainUUID, txHash)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			return err
		}
	}

	if minedJob != nil {
		err = uc.minedJob.Execute(ctx, minedJob)
		if err != nil {
			return err
		}

		err = uc.pendingJobState.Remove(ctx, minedJob.UUID)
		if err != nil {
			logger.WithError(err).Error("failed to remove pending job")
			return err
		}
	}

	return nil
}
