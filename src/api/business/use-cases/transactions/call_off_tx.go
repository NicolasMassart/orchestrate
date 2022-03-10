package transactions

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/entities"
)

const callOffTxComponent = "use-cases.call-off-tx"

type callOffTxUseCase struct {
	getTxUC      usecases.GetTxUseCase
	retryJobTxUC usecases.RetryJobTxUseCase
	logger       *log.Logger
}

func NewCallOffTxUseCase(getTxUC usecases.GetTxUseCase, retryJobTxUC usecases.RetryJobTxUseCase) usecases.CallOffTxUseCase {
	return &callOffTxUseCase{
		getTxUC:      getTxUC,
		retryJobTxUC: retryJobTxUC,
		logger:       log.NewLogger().SetComponent(callOffTxComponent),
	}
}

func (uc *callOffTxUseCase) Execute(ctx context.Context, scheduleUUID string, userInfo *multitenancy.UserInfo) (*entities.TxRequest, error) {
	ctx = log.WithFields(
		ctx,
		log.Field("schedule", scheduleUUID),
	)
	logger := uc.logger.WithContext(ctx)
	logger.Debug("calling off pending transaction")

	tx, err := uc.getTxUC.Execute(ctx, scheduleUUID, userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	if tx.Params.Protocol != "" {
		errMsg := "call off is not supported for private transaction"
		logger.Error(errMsg)
		return nil, errors.InvalidParameterError(errMsg).ExtendComponent(sendTxComponent)
	}

	if tx.InternalData != nil && tx.InternalData.OneTimeKey {
		errMsg := "call off is not supported for oneTimeKey transactions"
		logger.Error(errMsg)
		return nil, errors.InvalidParameterError(errMsg).ExtendComponent(sendTxComponent)
	}

	job := tx.Schedule.Jobs[0]
	err = uc.retryJobTxUC.Execute(ctx, job.UUID, 0.1, nil, userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	txRequest, err := uc.getTxUC.Execute(ctx, scheduleUUID, userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	logger.WithField("schedule", txRequest.Schedule.UUID).Info("cancel transaction was sent successfully")
	return txRequest, nil
}
