package usecases

import (
	"context"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -source=chains.go -destination=mocks/chains.go -package=mocks

type ChainBlock interface {
	Execute(ctx context.Context, chainUUID string, blockNumber uint64, txHashes []*ethcommon.Hash) error
}

type ChainUseCases interface {
	ChainBlockUseCase() ChainBlock
}
