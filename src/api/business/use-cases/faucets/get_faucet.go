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

const getFaucetComponent = "use-cases.get-faucet"

// getFaucetUseCase is a use case to get a faucet
type getFaucetUseCase struct {
	db     store.DB
	logger *log.Logger
}

// NewGetFaucetUseCase creates a new GetFaucetUseCase
func NewGetFaucetUseCase(db store.DB) usecases.GetFaucetUseCase {
	return &getFaucetUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(getFaucetComponent),
	}
}

// Execute gets a faucet
func (uc *getFaucetUseCase) Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) (*entities.Faucet, error) {
	ctx = log.WithFields(ctx, log.Field("faucet", uuid))
	logger := uc.logger.WithContext(ctx)

	faucet, err := uc.db.Faucet().FindOneByUUID(ctx, uuid, userInfo.AllowedTenants)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(getFaucetComponent)
	}

	logger.Debug("faucet found successfully")
	return faucet, nil
}
