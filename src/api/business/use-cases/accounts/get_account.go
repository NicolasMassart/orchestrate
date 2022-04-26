package accounts

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

const getAccountComponent = "use-cases.get-account"

type getAccountUseCase struct {
	db     store.DB
	logger *log.Logger
}

func NewGetAccountUseCase(db store.DB) usecases.GetAccountUseCase {
	return &getAccountUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(getAccountComponent),
	}
}

func (uc *getAccountUseCase) Execute(ctx context.Context, address ethcommon.Address, userInfo *multitenancy.UserInfo) (*entities.Account, error) {
	ctx = log.WithFields(ctx, log.Field("address", address))
	logger := uc.logger.WithContext(ctx)

	acc, err := uc.db.Account().FindOneByAddress(ctx, address.Hex(), userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(getAccountComponent)
	}

	logger.Trace("account found successfully")
	return acc, nil
}
