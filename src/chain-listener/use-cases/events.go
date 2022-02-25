package usecases

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -source=events.go -destination=mocks/events.go -package=mocks

type PendingJobEventHandler interface {
	Execute(ctx context.Context, job *entities.Job) error
}

type ChainBlockEventHandler interface {
	Execute(ctx context.Context, blockNumber uint64, txHashes []*ethcommon.Hash) error
}

type NewChainEventHandler interface {
	Execute(ctx context.Context, event *entities.Chain) error
}

type UpdateChainEventHandler interface {
	Execute(ctx context.Context, event *entities.Chain) error
}
type DeleteChainEventHandler interface {
	Execute(ctx context.Context, event *entities.Chain) error
}
