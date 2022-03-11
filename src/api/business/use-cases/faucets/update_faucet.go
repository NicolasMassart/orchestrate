package faucets

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
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

	if faucet.ChainRule != "" {
		_, err := uc.db.Chain().FindOneByUUID(ctx, faucet.ChainRule, userInfo.AllowedTenants, userInfo.Username)
		if errors.IsNotFoundError(err) {
			return nil, errors.InvalidParameterError("cannot find new linked chain")
		} else if err != nil {
			return nil, err
		}
	}

	if faucet.CreditorAccount.String() != new(ethcommon.Address).String() {
		_, err := uc.db.Account().FindOneByAddress(ctx, faucet.CreditorAccount.String(), userInfo.AllowedTenants, userInfo.Username)
		if errors.IsNotFoundError(err) {
			return nil, errors.InvalidParameterError("cannot find updated creditor account")
		} else if err != nil {
			return nil, err
		}
	}

	var err error
	faucet, err = uc.db.Faucet().Update(ctx, faucet, userInfo.AllowedTenants)
	if err != nil {
		return nil, err
	}

	faucet, err = uc.db.Faucet().FindOneByUUID(ctx, faucet.UUID, userInfo.AllowedTenants)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateFaucetComponent)
	}

	logger.Info("faucet updated successfully")
	return faucet, nil
}
