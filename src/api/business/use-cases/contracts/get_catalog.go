package contracts

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
)

const getCatalogComponent = "use-cases.get-catalog"

type getCatalogUseCase struct {
	agent  store.ContractAgent
	logger *log.Logger
}

func NewGetCatalogUseCase(agent store.ContractAgent) usecases.GetContractsCatalogUseCase {
	return &getCatalogUseCase{
		agent:  agent,
		logger: log.NewLogger().SetComponent(getCatalogComponent),
	}
}

// TODO: Modify to get all contracts and then only return necessary fields instead of getting only names
// Execute gets all contract names from Postgres
func (uc *getCatalogUseCase) Execute(ctx context.Context) ([]string, error) {
	names, err := uc.agent.ListNames(ctx)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(getCatalogComponent)
	}

	uc.logger.WithContext(ctx).Debug("get catalog executed successfully")
	return names, nil
}
