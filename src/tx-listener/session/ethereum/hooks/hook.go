package hook

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/src/tx-listener/dynamic"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

//go:generate mockgen -source=hook.go -destination=mock/mock.go -package=mock

type Hook interface {
	AfterNewBlock(ctx context.Context, chain *dynamic.Chain, block *ethtypes.Block, jobs []*entities.Job) error
}
