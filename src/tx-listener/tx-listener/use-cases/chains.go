package usecases

import (
	"context"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -source=chains.go -destination=mocks/chains.go -package=mocks

type ChainBlockTxs interface {
	Execute(ctx context.Context, chainUUID string, blockNumber uint64, txHashes []*ethcommon.Hash) error
}

type ChainBlockEvents interface {
	Execute(ctx context.Context, chainUUID string, blockNumber uint64) error
}

type ChainUseCases interface {
	ChainBlockTxsUseCase() ChainBlockTxs
	ChainBlockEventsUseCase() ChainBlockEvents
}
