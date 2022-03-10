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

const getEventsComponent = "use-cases.get-events"

type getEventsUseCase struct {
	agent  store.ContractEventAgent
	logger *log.Logger
}

func NewGetEventsUseCase(agent store.ContractEventAgent) usecases.GetContractEventsUseCase {
	return &getEventsUseCase{
		agent:  agent,
		logger: log.NewLogger().SetComponent(getEventsComponent),
	}
}

// Execute validates and registers a new contract in Postgres
func (uc *getEventsUseCase) Execute(ctx context.Context, chainID string, address ethcommon.Address, sigHash hexutil.Bytes, indexedInputCount uint32) (abi string, eventsABI []string, err error) {
	ctx = log.WithFields(ctx, log.Field("chain_id", chainID), log.Field("address", address))
	logger := uc.logger.WithContext(ctx)

	contractEvent, err := uc.agent.FindOneByAccountAndSigHash(ctx, chainID, address.Hex(), sigHash.String(), indexedInputCount)
	if err != nil && !errors.IsNotFoundError(err) {
		return "", nil, errors.FromError(err).ExtendComponent(getEventsComponent)
	}

	if contractEvent != nil {
		logger.Debug("events were fetched successfully")
		return contractEvent.ABI, nil, nil
	}

	defaultEventModels, err := uc.agent.FindDefaultBySigHash(ctx, sigHash.String(), indexedInputCount)
	if err != nil {
		return "", nil, errors.FromError(err).ExtendComponent(getEventsComponent)
	}

	for _, e := range defaultEventModels {
		eventsABI = append(eventsABI, e.ABI)
	}

	logger.Debug("default events were fetched successfully")
	return "", eventsABI, nil
}
