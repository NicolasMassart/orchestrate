package usecases

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

//go:generate mockgen -source=transactions.go -destination=mocks/transactions.go -package=mocks

type TransactionUseCases interface {
	SendContractTransaction() SendContractTxUseCase
	SendDeployTransaction() SendDeployTxUseCase
	SendTransaction() SendTxUseCase
	GetTransaction() GetTxUseCase
	SearchTransactions() SearchTransactionsUseCase
	SpeedUpTransaction() SpeedUpTxUseCase
	CallOffTransaction() CallOffTxUseCase
}

type GetTxUseCase interface {
	Execute(ctx context.Context, scheduleUUID string, userInfo *multitenancy.UserInfo) (*entities.TxRequest, error)
}

type SearchTransactionsUseCase interface {
	Execute(ctx context.Context, filters *entities.TransactionRequestFilters, userInfo *multitenancy.UserInfo) ([]*entities.TxRequest, error)
}

type SendDeployTxUseCase interface {
	Execute(ctx context.Context, txRequest *entities.TxRequest, userInfo *multitenancy.UserInfo) (*entities.TxRequest, error)
}
type SendContractTxUseCase interface {
	Execute(ctx context.Context, txRequest *entities.TxRequest, userInfo *multitenancy.UserInfo) (*entities.TxRequest, error)
}

type SendTxUseCase interface {
	Execute(ctx context.Context, txRequest *entities.TxRequest, txData hexutil.Bytes, userInfo *multitenancy.UserInfo) (*entities.TxRequest, error)
}

type SpeedUpTxUseCase interface {
	Execute(ctx context.Context, scheduleUUID string, gasIncrement float64, userInfo *multitenancy.UserInfo) (*entities.TxRequest, error)
}

type CallOffTxUseCase interface {
	Execute(ctx context.Context, scheduleUUID string, userInfo *multitenancy.UserInfo) (*entities.TxRequest, error)
}
