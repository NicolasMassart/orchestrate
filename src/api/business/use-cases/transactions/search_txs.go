package transactions

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
)

const searchTxsComponent = "use-cases.search-txs"

type searchTransactionsUseCase struct {
	db           store.DB
	getTxUseCase usecases.GetTxUseCase
	logger       *log.Logger
}

func NewSearchTransactionsUseCase(db store.DB, getTxUseCase usecases.GetTxUseCase) usecases.SearchTransactionsUseCase {
	return &searchTransactionsUseCase{
		db:           db,
		getTxUseCase: getTxUseCase,
		logger:       log.NewLogger().SetComponent(searchTxsComponent),
	}
}

func (uc *searchTransactionsUseCase) Execute(ctx context.Context, filters *entities.TransactionRequestFilters, userInfo *multitenancy.UserInfo) ([]*entities.TxRequest, error) {
	txRequests, err := uc.db.TransactionRequest().Search(ctx, filters, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(searchTxsComponent)
	}

	for idx, txRequest := range txRequests {
		tx, err := uc.getTxUseCase.Execute(ctx, txRequest.Schedule.UUID, userInfo)
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(searchTxsComponent)
		}
		txRequests[idx] = tx
	}

	uc.logger.Trace("transaction requests found successfully")
	return txRequests, nil
}
