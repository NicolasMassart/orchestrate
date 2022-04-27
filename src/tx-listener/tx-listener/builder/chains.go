package builder

import (
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/chains"
)

type chainUCs struct {
	chainBlockUseCase usecases.ChainBlock
}

func (s *chainUCs) ChainBlockUseCase() usecases.ChainBlock {
	return s.chainBlockUseCase
}

func NewChainUseCases(jobUCs usecases.JobUseCases,
	state store.State,
	logger *log.Logger,
) usecases.ChainUseCases {
	chainBlock := chains.NewChainBlockUseCase(jobUCs.MinedJobUseCase(), state.PendingJobState(), logger)
	return &chainUCs{
		chainBlockUseCase: chainBlock,
	}
}
