package events

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/chain-listener/state"
	usecases "github.com/consensys/orchestrate/src/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/entities"
)

const updateChainEventUseCaseComponent = "chain-listener.use-case.event.update-chain"

type updateChainEventHandler struct {
	chainState     state.Chain
	sessionHandler usecases.SessionHandler
	logger         *log.Logger
}

func UpdateChainEventHandler(sessionHandler usecases.SessionHandler,
	chainState state.Chain,
	logger *log.Logger,
) usecases.UpdateChainEventHandler {
	return &updateChainEventHandler{
		sessionHandler: sessionHandler,
		chainState:     chainState,
		logger:         logger.SetComponent(updateChainEventUseCaseComponent),
	}
}

func (uc *updateChainEventHandler) Execute(ctx context.Context, chain *entities.Chain) error {
	logger := uc.logger.WithField("chain", chain.UUID)
	err := uc.chainState.Update(ctx, chain)
	if err != nil {
		logger.WithError(err).Error("failed to update chain")
		return err
	}
	logger.Info("chain update successfully")
	return nil
}
