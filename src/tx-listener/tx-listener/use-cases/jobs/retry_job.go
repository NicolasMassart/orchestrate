package jobs

import (
	"context"
	"math"
	"math/big"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const retryJobUseCaseComponent = "tx-listener.use-case.retry-job"

type retryJobUseCase struct {
	client orchestrateclient.OrchestrateClient
	logger *log.Logger
}

func RetryJobUseCase(client orchestrateclient.OrchestrateClient, logger *log.Logger) usecases.RetryJob {
	return &retryJobUseCase{
		client: client,
		logger: logger.SetComponent(retryJobUseCaseComponent),
	}
}

// Execute starts a job session
func (uc *retryJobUseCase) Execute(ctx context.Context, job *entities.Job, childUUID string, nChildren int) (string, error) {
	logger := uc.logger.WithField("job", job.UUID).WithField("children", nChildren)
	logger.Debug("retrying job")

	// In case gas increments on every retry we create a new job
	if job.Type != entities.EthereumRawTransaction &&
		(job.InternalData.GasPriceIncrement > 0.0 &&
			nChildren <= int(math.Ceil(job.InternalData.GasPriceLimit/job.InternalData.GasPriceIncrement))) {

		childJob, errr := uc.createAndStartNewChildJob(ctx, job, nChildren)
		if errr != nil {
			return "", errors.FromError(errr).ExtendComponent(retryJobUseCaseComponent)
		}

		logger.WithField("child_job", childJob.UUID).Info("new child job created and started")
		return childJob.UUID, nil
	}

	// Otherwise we retry on last job
	err := uc.client.ResendJobTx(ctx, childUUID)
	if err != nil {
		logger.WithError(err).Error("failed to resend job")
		return "", errors.FromError(err).ExtendComponent(retryJobUseCaseComponent)
	}

	logger.Info("job has been resent successfully")
	return job.UUID, nil
}

func (uc *retryJobUseCase) createAndStartNewChildJob(ctx context.Context,
	parentJob *entities.Job,
	nChildrenJobs int,
) (*types.JobResponse, error) {
	logger := uc.logger.WithContext(ctx).WithField("job", parentJob.UUID)
	gasPriceMultiplier := getGasPriceMultiplier(
		parentJob.InternalData.GasPriceIncrement,
		parentJob.InternalData.GasPriceLimit,
		float64(nChildrenJobs),
	)

	childJobRequest := craftChildJobRequest(parentJob, gasPriceMultiplier)
	childJob, err := uc.client.CreateJob(ctx, childJobRequest)
	if err != nil {
		logger.Error("failed create new child job")
		return nil, errors.FromError(err).ExtendComponent(retryJobUseCaseComponent)
	}

	err = uc.client.StartJob(ctx, childJob.UUID)
	if err != nil {
		logger.WithField("child_job", childJob.UUID).Error("failed start child job")
		return nil, errors.FromError(err).ExtendComponent(retryJobUseCaseComponent)
	}

	return childJob, nil
}

func getGasPriceMultiplier(increment, limit, nChildren float64) float64 {
	// This is fine as GasPriceIncrement default value is 0
	newGasPriceMultiplier := (nChildren + 1) * increment

	if newGasPriceMultiplier >= limit {
		newGasPriceMultiplier = limit
	}

	return newGasPriceMultiplier
}

func craftChildJobRequest(parentJob *entities.Job, gasPriceMultiplier float64) *types.CreateJobRequest {
	// We selectively choose fields from the parent job
	newJobRequest := &types.CreateJobRequest{
		ChainUUID:     parentJob.ChainUUID,
		ScheduleUUID:  parentJob.ScheduleUUID,
		Type:          parentJob.Type,
		Labels:        parentJob.Labels,
		ParentJobUUID: parentJob.UUID,
	}

	// raw transactions are resent as-is with no modifications
	if parentJob.Type == entities.EthereumRawTransaction {
		newJobRequest.Transaction = types.ETHTransactionRequest{
			Raw: parentJob.Transaction.Raw,
		}

		return newJobRequest
	}

	newJobRequest.Transaction = types.ETHTransactionRequest{
		From:           parentJob.Transaction.From,
		To:             parentJob.Transaction.To,
		Value:          parentJob.Transaction.Value,
		Data:           parentJob.Transaction.Data,
		PrivateFrom:    parentJob.Transaction.PrivateFrom,
		PrivateFor:     parentJob.Transaction.PrivateFor,
		PrivacyGroupID: parentJob.Transaction.PrivacyGroupID,
		Nonce:          parentJob.Transaction.Nonce,
	}

	switch parentJob.Transaction.TransactionType {
	case entities.LegacyTxType:
		curGasPriceF := new(big.Float).SetInt(parentJob.Transaction.GasPrice.ToInt())
		nextGasPriceF := new(big.Float).Mul(curGasPriceF, big.NewFloat(1+gasPriceMultiplier))
		nextGasPrice := new(big.Int)
		nextGasPriceF.Int(nextGasPrice)
		newJobRequest.Transaction.GasPrice = (*hexutil.Big)(nextGasPrice)
	case entities.DynamicFeeTxType:
		curGasTipCapF := new(big.Float).SetInt(parentJob.Transaction.GasTipCap.ToInt())
		nextGasTipCapF := new(big.Float).Mul(curGasTipCapF, big.NewFloat(1+gasPriceMultiplier))
		nextGasTipCap := new(big.Int)
		nextGasTipCapF.Int(nextGasTipCap)
		newJobRequest.Transaction.GasTipCap = (*hexutil.Big)(nextGasTipCap)
	}

	return newJobRequest
}
