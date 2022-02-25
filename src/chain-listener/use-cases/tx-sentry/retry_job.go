package txsentry

import (
	"context"
	"math"
	"math/big"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/types"
	usecases "github.com/consensys/orchestrate/src/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const retrySessionJobComponent = "chain-listener.use-case.tx-sentry.retry-job"

type retryJobUseCase struct {
	client orchestrateclient.OrchestrateClient
	logger *log.Logger
}

func RetrySessionJobUseCase(client orchestrateclient.OrchestrateClient, logger *log.Logger) usecases.SendRetryJob {
	return &retryJobUseCase{
		client: client,
		logger: logger.SetComponent(retrySessionJobComponent),
	}
}

// Execute starts a job session
func (uc *retryJobUseCase) Execute(ctx context.Context, jobUUID, childUUID string, nChildren int) (string, error) {
	logger := uc.logger.WithField("job", jobUUID).WithField("children", nChildren)
	logger.Debug("retrying job")

	job, err := uc.client.GetJob(ctx, jobUUID)
	if err != nil {
		logger.WithError(err).Error("failed to get job")
		return "", errors.FromError(err).ExtendComponent(retrySessionJobComponent)
	}

	status := job.Status
	if status != entities.StatusPending {
		logger.WithField("status", status).Info("job has been updated. stopping job session")
		return "", nil
	}

	// In case gas increments on every retry we create a new job
	if job.Type != entities.EthereumRawTransaction &&
		(job.Annotations.GasPricePolicy.RetryPolicy.Increment > 0.0 &&
			nChildren <= int(math.Ceil(job.Annotations.GasPricePolicy.RetryPolicy.Limit/job.Annotations.GasPricePolicy.RetryPolicy.Increment))) {

		childJob, errr := uc.createAndStartNewChildJob(ctx, job, nChildren)
		if errr != nil {
			return "", errors.FromError(errr).ExtendComponent(retrySessionJobComponent)
		}

		logger.WithField("child_job", childJob.UUID).Info("new child job created and started")
		return childJob.UUID, nil
	}

	// Otherwise we retry on last job
	err = uc.client.ResendJobTx(ctx, childUUID)
	if err != nil {
		logger.WithError(err).Error("failed to resend job")
		return "", errors.FromError(err).ExtendComponent(retrySessionJobComponent)
	}

	logger.Info("job has been resent successfully")
	return job.UUID, nil
}

func (uc *retryJobUseCase) createAndStartNewChildJob(ctx context.Context,
	parentJob *types.JobResponse,
	nChildrenJobs int,
) (*types.JobResponse, error) {
	logger := uc.logger.WithContext(ctx).WithField("job", parentJob.UUID)
	gasPriceMultiplier := getGasPriceMultiplier(
		parentJob.Annotations.GasPricePolicy.RetryPolicy.Increment,
		parentJob.Annotations.GasPricePolicy.RetryPolicy.Limit,
		float64(nChildrenJobs),
	)

	childJobRequest := newChildJobRequest(parentJob, gasPriceMultiplier)
	childJob, err := uc.client.CreateJob(ctx, childJobRequest)
	if err != nil {
		logger.Error("failed create new child job")
		return nil, errors.FromError(err).ExtendComponent(retrySessionJobComponent)
	}

	err = uc.client.StartJob(ctx, childJob.UUID)
	if err != nil {
		logger.WithField("child_job", childJob.UUID).Error("failed start child job")
		return nil, errors.FromError(err).ExtendComponent(retrySessionJobComponent)
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

func newChildJobRequest(parentJob *types.JobResponse, gasPriceMultiplier float64) *types.CreateJobRequest {
	// We selectively choose fields from the parent job
	newJobRequest := &types.CreateJobRequest{
		ChainUUID:     parentJob.ChainUUID,
		ScheduleUUID:  parentJob.ScheduleUUID,
		Type:          parentJob.Type,
		Labels:        parentJob.Labels,
		Annotations:   parentJob.Annotations,
		ParentJobUUID: parentJob.UUID,
	}

	// raw transactions are resent as-is with no modifications
	if parentJob.Type == entities.EthereumRawTransaction {
		newJobRequest.Transaction = types.ETHTransactionRequest{
			Raw: utils.StringToHexBytes(parentJob.Transaction.Raw),
		}

		return newJobRequest
	}

	newJobRequest.Transaction = types.ETHTransactionRequest{
		From:           utils.ToPtr(ethcommon.HexToAddress(parentJob.Transaction.From)).(*ethcommon.Address),
		To:             utils.ToPtr(ethcommon.HexToAddress(parentJob.Transaction.To)).(*ethcommon.Address),
		Value:          utils.StringBigIntToHex(parentJob.Transaction.Value),
		Data:           utils.StringToHexBytes(parentJob.Transaction.Data),
		PrivateFrom:    parentJob.Transaction.PrivateFrom,
		PrivateFor:     parentJob.Transaction.PrivateFor,
		PrivacyGroupID: parentJob.Transaction.PrivacyGroupID,
		Nonce:          parentJob.Transaction.Nonce,
	}

	switch parentJob.Transaction.TransactionType {
	case entities.LegacyTxType.String():
		curGasPriceF, _ := new(big.Float).SetString(parentJob.Transaction.GasPrice)
		nextGasPriceF := new(big.Float).Mul(curGasPriceF, big.NewFloat(1+gasPriceMultiplier))
		nextGasPrice := new(big.Int)
		nextGasPriceF.Int(nextGasPrice)
		newJobRequest.Transaction.GasPrice = (*hexutil.Big)(nextGasPrice)
	case entities.DynamicFeeTxType.String():
		curGasTipCapF, _ := new(big.Float).SetString(parentJob.Transaction.GasTipCap)
		nextGasTipCapF := new(big.Float).Mul(curGasTipCapF, big.NewFloat(1+gasPriceMultiplier))
		nextGasTipCap := new(big.Int)
		nextGasTipCapF.Int(nextGasTipCap)
		newJobRequest.Transaction.GasTipCap = (*hexutil.Big)(nextGasTipCap)
	}

	return newJobRequest
}
