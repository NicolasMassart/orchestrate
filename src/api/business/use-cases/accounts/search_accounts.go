package accounts

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
)

const searchAccountsComponent = "use-cases.search-accounts"

type searchAccountsUseCase struct {
	db     store.DB
	logger *log.Logger
}

func NewSearchAccountsUseCase(db store.DB) usecases.SearchAccountsUseCase {
	return &searchAccountsUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(searchAccountsComponent),
	}
}

func (uc *searchAccountsUseCase) Execute(ctx context.Context, filters *entities.AccountFilters, userInfo *multitenancy.UserInfo) ([]*entities.Account, error) {
	accs, err := uc.db.Account().Search(ctx, filters, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(searchAccountsComponent)
	}

	uc.logger.WithContext(ctx).Trace("accounts found successfully")
	return accs, nil
}
