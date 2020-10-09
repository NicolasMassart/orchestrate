package transactions

import (
	"context"

	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types/entities"
	usecases "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/transaction-scheduler/use-cases"

	log "github.com/sirupsen/logrus"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/errors"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/store"
)

const getTxComponent = "use-cases.get-tx"

// getTxUseCase is a use case to get a transaction request
type getTxUseCase struct {
	db                 store.DB
	getScheduleUsecase usecases.GetScheduleUseCase
}

// NewGetTxUseCase creates a new GetTxUseCase
func NewGetTxUseCase(db store.DB, getScheduleUsecase usecases.GetScheduleUseCase) usecases.GetTxUseCase {
	return &getTxUseCase{
		db:                 db,
		getScheduleUsecase: getScheduleUsecase,
	}
}

// Execute gets a transaction request
func (uc *getTxUseCase) Execute(ctx context.Context, scheduleUUID string, tenants []string) (*entities.TxRequest, error) {
	logger := log.WithContext(ctx).WithField("tx_request_uuid", scheduleUUID)
	logger.Debug("getting transaction request")

	txRequestModel, err := uc.db.TransactionRequest().FindOneByUUID(ctx, scheduleUUID, tenants)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(getTxComponent)
	}

	txRequest := &entities.TxRequest{
		IdempotencyKey: txRequestModel.IdempotencyKey,
		ChainName:      txRequestModel.ChainName,
		Params:         txRequestModel.Params,
		CreatedAt:      txRequestModel.CreatedAt,
	}
	txRequest.Schedule, err = uc.getScheduleUsecase.Execute(ctx, txRequestModel.Schedule.UUID, tenants)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(getTxComponent)
	}

	logger.Info("transaction request found successfully")

	return txRequest, nil
}
