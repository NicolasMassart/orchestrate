package transactions

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/entities"
)

const speedUpTxComponent = "use-cases.speed-up-tx"

type speedUpTxUseCase struct {
	getTxUC      usecases.GetTxUseCase
	retryJobTxUC usecases.RetryJobTxUseCase
	logger       *log.Logger
}

func NewSpeedUpTxUseCase(getTxUC usecases.GetTxUseCase, retryJobTxUC usecases.RetryJobTxUseCase) usecases.SpeedUpTxUseCase {
	return &speedUpTxUseCase{
		getTxUC:      getTxUC,
		retryJobTxUC: retryJobTxUC,
		logger:       log.NewLogger().SetComponent(speedUpTxComponent),
	}
}

func (uc *speedUpTxUseCase) Execute(ctx context.Context, scheduleUUID string, gasIncrement float64, userInfo *multitenancy.UserInfo) (*entities.TxRequest, error) {
	ctx = log.WithFields(
		ctx,
		log.Field("schedule", scheduleUUID),
	)
	logger := uc.logger.WithContext(ctx)
	logger.Debug("speeding up transaction")

	tx, err := uc.getTxUC.Execute(ctx, scheduleUUID, userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	if tx.Params.Protocol != "" {
		errMsg := "speed up is not supported for private transactions"
		logger.Error(errMsg)
		return nil, errors.InvalidParameterError(errMsg).ExtendComponent(sendTxComponent)
	}

	if tx.InternalData != nil && tx.InternalData.OneTimeKey {
		errMsg := "speed up is not supported for oneTimeKey transactions"
		logger.Error(errMsg)
		return nil, errors.InvalidParameterError(errMsg).ExtendComponent(sendTxComponent)
	}

	job := tx.Schedule.Jobs[0]
	err = uc.retryJobTxUC.Execute(ctx, job.UUID, gasIncrement, job.Transaction.Data, userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	txRequest, err := uc.getTxUC.Execute(ctx, scheduleUUID, userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	logger.WithField("schedule", txRequest.Schedule.UUID).Info("speed-up transaction was sent successfully")
	return txRequest, nil
}
