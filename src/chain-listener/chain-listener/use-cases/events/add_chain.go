package events

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/chain-listener/store"
	"github.com/consensys/orchestrate/src/entities"
)

const newChainEventUseCaseComponent = "chain-listener.use-case.event.add-chain"

type addChainUC struct {
	chainState store.Chain
	logger     *log.Logger
}

func AddChainUseCase(chainState store.Chain,
	logger *log.Logger,
) usecases.AddChainUseCase {
	return &addChainUC{
		chainState: chainState,
		logger:     logger.SetComponent(newChainEventUseCaseComponent),
	}
}

func (uc *addChainUC) Execute(ctx context.Context, chain *entities.Chain) error {
	logger := uc.logger.WithField("chain", chain.UUID)
	err := uc.chainState.Add(ctx, chain)
	if err != nil {
		logger.WithError(err).Error("failed to persist chain")
		return err
	}
	logger.Info("chain added to state successfully")
	return nil
}
