package events

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/chain-listener/state"
	usecases "github.com/consensys/orchestrate/src/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/entities"
)

const deleteChainEventUseCaseComponent = "chain-listener.use-case.event.delete-chain"

type deleteChainEventHandler struct {
	chainState        state.Chain
	sessionHandler    usecases.SessionHandler
	pendingJobState   state.PendingJob
	retrySessionState state.RetrySessions
	logger            *log.Logger
}

func DeleteChainEventHandler(sessionHandler usecases.SessionHandler,
	chainState state.Chain,
	pendingJobState state.PendingJob,
	retrySessionState state.RetrySessions,
	logger *log.Logger,
) usecases.DeleteChainEventHandler {
	return &deleteChainEventHandler{
		sessionHandler:    sessionHandler,
		chainState:        chainState,
		pendingJobState:   pendingJobState,
		retrySessionState: retrySessionState,
		logger:            logger.SetComponent(deleteChainEventUseCaseComponent),
	}
}

func (uc *deleteChainEventHandler) Execute(ctx context.Context, chain *entities.Chain) error {
	logger := uc.logger.WithField("chain", chain.UUID)
	err := uc.pendingJobState.DeletePerChainUUID(ctx, chain.UUID)
	if err != nil && !errors.IsNotFoundError(err) {
		logger.WithError(err).Error("failed to delete all pending jobs per chain")
		return err
	}

	sessIDs, err := uc.retrySessionState.ListByChainUUID(ctx, chain.UUID)
	if err != nil && !errors.IsNotFoundError(err) {
		logger.WithError(err).Error("failed to retrieve all active retry job sessions per chain")
		return err
	}
	for _, sessID := range sessIDs {
		if err2 := uc.sessionHandler.StopSession(ctx, sessID); err2 != nil {
			return err2
		}
		if err2 := uc.retrySessionState.Remove(ctx, sessID); err2 != nil {
			return err2
		}
	}
	err = uc.chainState.Remove(ctx, chain.UUID)
	if err != nil {
		logger.WithError(err).Error("failed to delete chain")
		return err
	}
	logger.Info("chain deleted successfully")
	return nil
}
