package chains

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/infra/database"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
)

const deleteChainComponent = "use-cases.delete-chain"

// deleteChainUseCase is a use case to delete a chain
type deleteChainUseCase struct {
	db         store.DB
	getChainUC usecases.GetChainUseCase
	logger     *log.Logger
}

// NewDeleteChainUseCase creates a new DeleteChainUseCase
func NewDeleteChainUseCase(db store.DB, getChainUC usecases.GetChainUseCase) usecases.DeleteChainUseCase {
	return &deleteChainUseCase{
		db:         db,
		getChainUC: getChainUC,
		logger:     log.NewLogger().SetComponent(deleteChainComponent),
	}
}

// Execute deletes a chain
func (uc *deleteChainUseCase) Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) error {
	ctx = log.WithFields(ctx, log.Field("chain", uuid))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("deleting chain")

	chain, err := uc.getChainUC.Execute(ctx, uuid, userInfo)
	if err != nil {
		return errors.FromError(err).ExtendComponent(deleteChainComponent)
	}

	err = database.ExecuteInDBTx(uc.db, func(tx database.Tx) error {
		var der error
		if chain.PrivateTxManager != nil {
			der = tx.(store.Tx).PrivateTxManager().Delete(ctx, chain.PrivateTxManager.UUID)
			if der != nil {
				return der
			}
		}

		der = tx.(store.Tx).Chain().Delete(ctx, chain, userInfo.AllowedTenants)
		if der != nil {
			return der
		}

		return nil
	})

	if err != nil {
		return errors.FromError(err).ExtendComponent(updateChainComponent)
	}

	logger.Info("chain deleted successfully")
	return nil
}
