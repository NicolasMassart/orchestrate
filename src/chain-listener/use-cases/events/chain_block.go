package events

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/chain-listener/state"
	usecases "github.com/consensys/orchestrate/src/chain-listener/use-cases"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const chainBlockEventUseCaseComponent = "chain-listener.use-case.event.chain-block"

type chainBlockEventHandler struct {
	pendingJobState   state.PendingJob
	retrySessionState state.RetrySessions
	updateJobStatus   usecases.UpdatedJobStatus
	sessionHandler    usecases.SessionHandler
	logger            *log.Logger
}

func ChainBlockEventHandler(updateJobStatus usecases.UpdatedJobStatus,
	sessionHandler usecases.SessionHandler,
	pendingJobState state.PendingJob,
	retrySessionState state.RetrySessions,
	logger *log.Logger,
) usecases.ChainBlockEventHandler {
	return &chainBlockEventHandler{
		pendingJobState:   pendingJobState,
		retrySessionState: retrySessionState,
		sessionHandler:    sessionHandler,
		updateJobStatus:   updateJobStatus,
		logger:            logger.SetComponent(chainBlockEventUseCaseComponent),
	}
}

func (uc *chainBlockEventHandler) Execute(ctx context.Context, blockNumber uint64, txHashes []*ethcommon.Hash) error {
	logger := uc.logger.WithField("block", blockNumber)
	logger.WithField("txs", len(txHashes)).Debug("processing block")

	// @TODO Run in parallel
	for _, txHash := range txHashes {
		if err := uc.handlePendingJob(ctx, txHash); err != nil {
			return err
		}

		if err := uc.handleRetrySessionJob(ctx, txHash); err != nil {
			return err
		}
	}

	return nil
}

func (uc *chainBlockEventHandler) handlePendingJob(ctx context.Context, txHash *ethcommon.Hash) error {
	minedJob, err := uc.pendingJobState.GetByTxHash(ctx, txHash)
	if err != nil && errors.IsNotFoundError(err) {
		return nil
	} else if err != nil {
		return err
	}

	err = uc.updateJobStatus.Execute(ctx, minedJob)
	if err != nil {
		return err
	}
	err = uc.pendingJobState.Remove(ctx, minedJob.UUID)
	if err != nil {
		return err
	}

	return nil
}

func (uc *chainBlockEventHandler) handleRetrySessionJob(ctx context.Context, txHash *ethcommon.Hash) error {
	retryJobSessID, err := uc.retrySessionState.SearchByTxHash(ctx, txHash)
	if err != nil && errors.IsNotFoundError(err) {
		return nil
	} else if err != nil {
		return err
	}

	err = uc.sessionHandler.StopSession(ctx, retryJobSessID)
	if err != nil {
		return err
	}

	return nil
}
