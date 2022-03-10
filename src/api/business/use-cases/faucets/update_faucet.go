package faucets

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
)

const updateFaucetComponent = "use-cases.update-faucet"

type updateFaucetUseCase struct {
	db     store.DB
	logger *log.Logger
}

func NewUpdateFaucetUseCase(db store.DB) usecases.UpdateFaucetUseCase {
	return &updateFaucetUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(updateFaucetComponent),
	}
}

func (uc *updateFaucetUseCase) Execute(ctx context.Context, faucet *entities.Faucet, userInfo *multitenancy.UserInfo) (*entities.Faucet, error) {
	ctx = log.WithFields(ctx, log.Field("faucet", faucet.UUID))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("updating faucet")

	_, err := uc.db.Faucet().Update(ctx, faucet, userInfo.AllowedTenants)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateFaucetComponent)
	}

	faucet, err = uc.db.Faucet().FindOneByUUID(ctx, faucet.UUID, userInfo.AllowedTenants)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateFaucetComponent)
	}

	logger.Info("faucet updated successfully")
	return faucet, nil
}
