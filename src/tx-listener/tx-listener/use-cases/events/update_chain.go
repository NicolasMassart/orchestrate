package events

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const updateChainUseCaseComponent = "tx-listener.use-case.event.update-chain"

type updateChainUC struct {
	chainState store.Chain
	logger     *log.Logger
}

func UpdateChainUseCase(chainState store.Chain,
	logger *log.Logger,
) usecases.UpdateChainUseCase {
	return &updateChainUC{
		chainState: chainState,
		logger:     logger.SetComponent(updateChainUseCaseComponent),
	}
}

func (uc *updateChainUC) Execute(ctx context.Context, chain *entities.Chain) error {
	logger := uc.logger.WithField("chain", chain.UUID)
	err := uc.chainState.Update(ctx, chain)
	if err != nil {
		if errors.IsNotFoundError(err) {
			logger.Warn(err)
			return nil
		}
		logger.WithError(err).Error("failed to update chain")
		return err
	}
	logger.Info("chain update successfully")
	return nil
}
