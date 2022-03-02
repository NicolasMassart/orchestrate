package events

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/chain-listener/store"
	"github.com/consensys/orchestrate/src/entities"
)

const deleteChainUseCaseComponent = "chain-listener.use-case.event.delete-chain"

type deleteChainUC struct {
	chainState        store.Chain
	sessionHandler    usecases.RetryJobSessionManager
	pendingJobState   store.PendingJob
	retrySessionState store.RetrySessions
	logger            *log.Logger
}

func DeleteChainUseCase(sessionHandler usecases.RetryJobSessionManager,
	chainState store.Chain,
	pendingJobState store.PendingJob,
	retrySessionState store.RetrySessions,
	logger *log.Logger,
) usecases.DeleteChainUseCase {
	return &deleteChainUC{
		sessionHandler:    sessionHandler,
		chainState:        chainState,
		pendingJobState:   pendingJobState,
		retrySessionState: retrySessionState,
		logger:            logger.SetComponent(deleteChainUseCaseComponent),
	}
}

func (uc *deleteChainUC) Execute(ctx context.Context, chain *entities.Chain) error {
	logger := uc.logger.WithField("chain", chain.UUID)
	sessIDs, err := uc.retrySessionState.ListByChainUUID(ctx, chain.UUID)
	if err != nil && !errors.IsNotFoundError(err) {
		logger.WithError(err).Error("failed to retrieve all active retry job sessions per chain")
		return err
	}
	for _, sessID := range sessIDs {
		if err2 := uc.sessionHandler.StopSession(ctx, sessID); err2 != nil {
			return err2
		}
	}

	err = uc.pendingJobState.DeletePerChainUUID(ctx, chain.UUID)
	if err != nil && !errors.IsNotFoundError(err) {
		logger.WithError(err).Error("failed to delete all pending jobs per chain")
		return err
	}

	err = uc.chainState.Remove(ctx, chain.UUID)
	if err != nil {
		logger.WithError(err).Error("failed to delete chain")
		return err
	}
	logger.Info("chain deleted successfully")
	return nil
}
