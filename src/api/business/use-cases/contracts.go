package usecases

import (
	"context"

	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

//go:generate mockgen -source=contracts.go -destination=mocks/contracts.go -package=mocks

type ContractUseCases interface {
	GetCatalog() GetContractsCatalogUseCase
	Get() GetContractUseCase
	GetContractEvents() GetContractEventsUseCase
	GetTags() GetContractTagsUseCase
	SetCodeHash() RegisterContractDeploymentUseCase
	Register() RegisterContractUseCase
	Search() SearchContractUseCase
	DecodeLog() DecodeEventLogUseCase
}

type GetContractsCatalogUseCase interface {
	Execute(ctx context.Context) ([]string, error)
}

type GetContractUseCase interface {
	Execute(ctx context.Context, name, tag string) (*entities.Contract, error)
}

type SearchContractUseCase interface {
	Execute(ctx context.Context, codehash hexutil.Bytes, address *ethcommon.Address) (*entities.Contract, error)
}

type GetContractEventsUseCase interface {
	Execute(ctx context.Context, chainID string, address ethcommon.Address, codeHash hexutil.Bytes, indexedInputCount uint32) (abi string, eventsABI []string, err error)
}

type GetContractTagsUseCase interface {
	Execute(ctx context.Context, name string) ([]string, error)
}

type RegisterContractUseCase interface {
	Execute(ctx context.Context, contract *entities.Contract) error
}

type RegisterContractDeploymentUseCase interface {
	Execute(ctx context.Context, chainID string, address ethcommon.Address, codeHash hexutil.Bytes) error
}

type DecodeEventLogUseCase interface {
	Execute(ctx context.Context, chainUUID string, eventLog *ethereum.Log) (*ethereum.Log, error)
}
