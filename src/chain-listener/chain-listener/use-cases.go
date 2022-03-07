package chainlistener

import (
	usecases "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases"
)

type EventUseCases interface {
	PendingJobUseCase() usecases.PendingJobUseCase
	DeleteChainUseCase() usecases.DeleteChainUseCase
	UpdateChainUseCase() usecases.UpdateChainUseCase
	AddChainUseCase() usecases.AddChainUseCase
	ChainBlockTxsUseCase() usecases.ChainBlockTxsUseCase
}
