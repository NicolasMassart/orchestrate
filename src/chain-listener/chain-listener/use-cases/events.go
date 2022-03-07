package usecases

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -source=events.go -destination=mocks/events.go -package=mocks

type PendingJobUseCase interface {
	Execute(ctx context.Context, job *entities.Job) error
}

type ChainBlockTxsUseCase interface {
	Execute(ctx context.Context, chainUUID string, blockNumber uint64, txHashes []*ethcommon.Hash) error
}

type AddChainUseCase interface {
	Execute(ctx context.Context, event *entities.Chain) error
}

type UpdateChainUseCase interface {
	Execute(ctx context.Context, event *entities.Chain) error
}
type DeleteChainUseCase interface {
	Execute(ctx context.Context, chainUUID string) error
}
