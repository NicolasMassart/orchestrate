package events

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const deleteChainUseCaseComponent = "tx-listener.use-case.event.delete-chain"

type deleteChainUC struct {
	chainState      store.Chain
	sessionHandler  usecases.RetryJobSessionManager
	pendingJobState store.PendingJob
	logger          *log.Logger
}

func DeleteChainUseCase(sessionHandler usecases.RetryJobSessionManager,
	chainState store.Chain,
	pendingJobState store.PendingJob,
	logger *log.Logger,
) usecases.DeleteChainUseCase {
	return &deleteChainUC{
		sessionHandler:  sessionHandler,
		chainState:      chainState,
		pendingJobState: pendingJobState,
		logger:          logger.SetComponent(deleteChainUseCaseComponent),
	}
}

func (uc *deleteChainUC) Execute(ctx context.Context, chainUUID string) error {
	logger := uc.logger.WithField("chain", chainUUID)
	if err := uc.sessionHandler.StopChainSessions(ctx, chainUUID); err != nil {
		return err
	}

	err := uc.pendingJobState.DeletePerChainUUID(ctx, chainUUID)
	if err != nil && !errors.IsNotFoundError(err) {
		logger.WithError(err).Error("failed to delete all pending jobs per chain")
		return err
	}

	err = uc.chainState.Delete(ctx, chainUUID)
	if err != nil {
		logger.WithError(err).Error("failed to delete chain")
		return err
	}
	logger.Info("chain deleted successfully")
	return nil
}
