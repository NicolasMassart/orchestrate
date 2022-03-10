package chains

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
)

const updateChainComponent = "use-cases.update-chain"

// updateChainUseCase is a use case to update a faucet
type updateChainUseCase struct {
	db         store.DB
	getChainUC usecases.GetChainUseCase
	logger     *log.Logger
}

// NewUpdateChainUseCase creates a new UpdateChainUseCase
func NewUpdateChainUseCase(db store.DB, getChainUC usecases.GetChainUseCase) usecases.UpdateChainUseCase {
	return &updateChainUseCase{
		db:         db,
		getChainUC: getChainUC,
		logger:     log.NewLogger().SetComponent(updateChainComponent),
	}
}

// Execute updates a chain
func (uc *updateChainUseCase) Execute(ctx context.Context, nextChain *entities.Chain, userInfo *multitenancy.UserInfo) (*entities.Chain, error) {
	ctx = log.WithFields(ctx, log.Field("chain", nextChain.UUID))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("updating chain")

	err := uc.db.Chain().Update(ctx, nextChain, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateChainComponent)
	}

	chain, err := uc.getChainUC.Execute(ctx, nextChain.UUID, userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateChainComponent)
	}

	logger.WithField("block", chain.ListenerCurrentBlock).
		Info("chain updated successfully")
	return chain, nil
}
