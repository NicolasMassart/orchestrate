package contracts

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
)

const getContractComponent = "use-cases.get-contract"

type getContractUseCase struct {
	agent  store.ContractAgent
	logger *log.Logger
}

func NewGetContractUseCase(agent store.ContractAgent) usecases.GetContractUseCase {
	return &getContractUseCase{
		agent:  agent,
		logger: log.NewLogger().SetComponent(getContractComponent),
	}
}

// Execute gets a contract from DB
func (uc *getContractUseCase) Execute(ctx context.Context, name, tag string) (*entities.Contract, error) {
	ctx = log.WithFields(ctx, log.Field("contract_name", name), log.Field("contract_tag", name))
	logger := uc.logger.WithContext(ctx)

	contract, err := uc.agent.FindOneByNameAndTag(ctx, name, tag)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(getContractComponent)
	}

	contract.ABI, err = abi.JSON(strings.NewReader(contract.RawABI))
	if err != nil {
		errMessage := "failed to parse contract abi"
		uc.logger.WithError(err).Error(errMessage)
		return nil, errors.DataCorruptedError(errMessage).ExtendComponent(getContractComponent)
	}

	// nolint
	for _, method := range contract.ABI.Methods {
		contract.Methods = append(contract.Methods, entities.ABIComponent{
			Signature: method.Sig,
		})
	}

	// nolint
	for _, event := range contract.ABI.Events {
		contract.Events = append(contract.Events, entities.ABIComponent{
			Signature: event.Sig,
		})
	}

	contract.Constructor = entities.ABIComponent{Signature: contract.ABI.Constructor.Sig}

	logger.Debug("contract found successfully")
	return contract, nil
}
