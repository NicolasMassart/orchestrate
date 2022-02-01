package contracts

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const registerDeploymentComponent = "use-cases.register-contract-deployment"

type registerDeploymentUseCase struct {
	agent  store.ContractAgent
	logger *log.Logger
}

func NewRegisterDeploymentUseCase(agent store.ContractAgent) usecases.RegisterContractDeploymentUseCase {
	return &registerDeploymentUseCase{
		agent:  agent,
		logger: log.NewLogger().SetComponent(registerDeploymentComponent),
	}
}

func (uc *registerDeploymentUseCase) Execute(ctx context.Context, chainID string, address ethcommon.Address, codeHash hexutil.Bytes) error {
	ctx = log.WithFields(ctx, log.Field("chain_id", chainID), log.Field("address", chainID))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("registering contract deployment hash is starting ...")

	err := uc.agent.RegisterDeployment(ctx, chainID, address, codeHash)
	if err != nil {
		return errors.FromError(err).ExtendComponent(registerDeploymentComponent)
	}

	logger.Debug("contract deployment is registered successfully")
	return nil
}
