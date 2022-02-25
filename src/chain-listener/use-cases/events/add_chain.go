package events

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/chain-listener/state"
	usecases "github.com/consensys/orchestrate/src/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/entities"
)

const newChainEventUseCaseComponent = "chain-listener.use-case.event.add-chain"

type addChainEventHandler struct {
	chainState     state.Chain
	sessionHandler usecases.SessionHandler
	logger         *log.Logger
}

func AddChainEventHandler(sessionHandler usecases.SessionHandler,
	chainState state.Chain,
	logger *log.Logger,
) usecases.NewChainEventHandler {
	return &addChainEventHandler{
		sessionHandler: sessionHandler,
		chainState:     chainState,
		logger:         logger.SetComponent(newChainEventUseCaseComponent),
	}
}

func (uc *addChainEventHandler) Execute(ctx context.Context, chain *entities.Chain) error {
	logger := uc.logger.WithField("chain", chain.UUID)
	err := uc.chainState.Add(ctx, chain)
	if err != nil {
		logger.WithError(err).Error("failed to persist chain")
		return err
	}
	logger.Info("chain added to state successfully")
	return nil
}
