package chains

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/parsers"
	"github.com/consensys/orchestrate/src/entities"
)

const searchChainsComponent = "use-cases.search-chains"

// searchChainsUseCase is a use case to search chains
type searchChainsUseCase struct {
	db     store.DB
	logger *log.Logger
}

// NewSearchChainsUseCase creates a new SearchChainsUseCase
func NewSearchChainsUseCase(db store.DB) usecases.SearchChainsUseCase {
	return &searchChainsUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(searchChainsComponent),
	}
}

// Execute search faucets
func (uc *searchChainsUseCase) Execute(ctx context.Context, filters *entities.ChainFilters, userInfo *multitenancy.UserInfo) ([]*entities.Chain, error) {
	logger := uc.logger.WithContext(ctx)

	chainModels, err := uc.db.Chain().Search(ctx, filters, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(searchChainsComponent)
	}

	var chains []*entities.Chain
	for _, chainModel := range chainModels {
		chains = append(chains, parsers.NewChainFromModel(chainModel))
	}

	logger.Debug("chains found successfully")
	return chains, nil
}
