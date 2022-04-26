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

const searchFaucetsComponent = "use-cases.search-faucets"

// searchFaucetsUseCase is a use case to search faucets
type searchFaucetsUseCase struct {
	db     store.DB
	logger *log.Logger
}

// NewSearchFaucets creates a new SearchFaucetsUseCase
func NewSearchFaucets(db store.DB) usecases.SearchFaucetsUseCase {
	return &searchFaucetsUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(searchFaucetsComponent),
	}
}

// Execute search faucets
func (uc *searchFaucetsUseCase) Execute(ctx context.Context, filters *entities.FaucetFilters, userInfo *multitenancy.UserInfo) ([]*entities.Faucet, error) {
	faucets, err := uc.db.Faucet().Search(ctx, filters, userInfo.AllowedTenants)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(searchFaucetsComponent)
	}

	uc.logger.Trace("faucets found successfully")
	return faucets, nil
}
