package builder

import (
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/contracts"
)

type contractsUCs struct {
	registerDeployedContractUseCase usecases.RegisterDeployedContract
}

func (s *contractsUCs) RegisterDeployedContractUseCase() usecases.RegisterDeployedContract {
	return s.registerDeployedContractUseCase
}

func NewContractUseCases(apiClient orchestrateclient.OrchestrateClient,
	ethClient ethclient.MultiClient,
	state store.State,
	logger *log.Logger,
) usecases.ContractsUseCases {
	registerDeployedContractUC := contracts.RegisterDeployedContractUseCase(apiClient, ethClient, state.ChainState(), logger)
	return &contractsUCs{
		registerDeployedContractUseCase: registerDeployedContractUC,
	}
}
