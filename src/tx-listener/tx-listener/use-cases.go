package txlistener

import (
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

type EventUseCases interface {
	PendingJobUseCase() usecases.PendingJobUseCase
	DeleteChainUseCase() usecases.DeleteChainUseCase
	UpdateChainUseCase() usecases.UpdateChainUseCase
	AddChainUseCase() usecases.AddChainUseCase
	ChainBlockTxsUseCase() usecases.ChainBlockTxsUseCase
}
