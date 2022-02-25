package txlistener

import (
	"context"
	"fmt"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/src/api/service/types"
	usecases "github.com/consensys/orchestrate/src/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const setMinedTxComponent = "chain-listener.use-case.tx-listener.updated-mined-job"

type updatedJobStatusUseCase struct {
	client           orchestrateclient.OrchestrateClient
	ethClient        ethclient.MultiClient
	sendNotification usecases.SendNotification
	logger           *log.Logger
}

func UpdateJobStatusUseCase(client orchestrateclient.OrchestrateClient,
	ethClient ethclient.MultiClient,
	sendNotification usecases.SendNotification,
	logger *log.Logger,
) usecases.UpdatedJobStatus {
	return &updatedJobStatusUseCase{
		client:           client,
		ethClient:        ethClient,
		sendNotification: sendNotification,
		logger:           logger.SetComponent(setMinedTxComponent),
	}
}

func (uc *updatedJobStatusUseCase) Execute(ctx context.Context, job *entities.Job) error {
	logger := uc.logger.WithField("job", job.UUID).WithField("tx_hash", job.Transaction.Hash)
	logger.Debug("updating job to mined")
	_, err := uc.getEthereumTx(ctx, job)
	if err != nil {
		return err
	}

	if job.Receipt == nil {
		job.Receipt, err = uc.getTxReceipt(ctx, job)
		if err != nil {
			return err
		}
	}

	// @TODO Decode logs
	err = uc.updateJobStatus(ctx, job, job.Receipt)
	if err != nil {
		return err
	}

	logger.Info("job was updated to mined successfully")

	err = uc.sendNotification.Execute(ctx, job)
	if err != nil {
		return err
	}

	return nil
}

func (uc *updatedJobStatusUseCase) updateJobStatus(ctx context.Context, job *entities.Job, receipt *ethereum.Receipt) error {

	updateTxReq := &types.UpdateJobRequest{
		Status:  entities.StatusMined,
		Message: fmt.Sprintf("transaction mined in block %v", receipt.BlockNumber),
	}

	if job.Transaction.TransactionType == entities.DynamicFeeTxType {
		effectiveGas, _ := hexutil.DecodeBig(receipt.EffectiveGasPrice)
		updateTxReq.Transaction = &types.ETHTransactionRequest{
			GasPrice: (*hexutil.Big)(effectiveGas),
		}
	}

	_, err := uc.client.UpdateJob(ctx, job.UUID, updateTxReq)
	if err != nil {
		errMsg := "failed to update job to MINED"
		uc.logger.WithError(err).Error(errMsg)
		return errors.DependencyFailureError(errMsg)
	}

	return nil
}

func (uc *updatedJobStatusUseCase) getEthereumTx(ctx context.Context, job *entities.Job) (*ethtypes.Transaction, error) {
	logger := uc.logger.WithField("job", job.UUID).WithField("tx_hash", job.Transaction.Hash)
	logger.Debug("fetching ethereum transaction data")

	chainURL := uc.client.ChainProxyURL(job.ChainUUID)
	ethTx, isPending, err := uc.ethClient.TransactionByHash(ctx, chainURL, *job.Transaction.Hash)
	if err != nil {
		errMsg := "failed to fetch transaction by hash"
		uc.logger.WithError(err).Error(errMsg)
		return nil, errors.DependencyFailureError(errMsg)
	}
	if isPending {
		errMsg := "cannot retrieve receipt from pending transaction"
		uc.logger.Error(errMsg)
		return nil, errors.InvalidStateError(errMsg)
	}

	return ethTx, nil
}

func (uc *updatedJobStatusUseCase) getTxReceipt(ctx context.Context, job *entities.Job) (*ethereum.Receipt, error) {
	logger := uc.logger.WithField("job", job.UUID).WithField("tx_hash", job.Transaction.Hash)
	chainURL := uc.client.ChainProxyURL(job.ChainUUID)

	switch job.Type {
	case entities.EEAMarkingTransaction:
		return uc.fetchPrivateReceipt(ctx, chainURL, *job.Transaction.Hash, logger)
	default:
		return uc.fetchReceipt(ctx, chainURL, *job.Transaction.Hash, logger)
	}

}

func (uc *updatedJobStatusUseCase) fetchReceipt(ctx context.Context, chainURL string, txHash ethcommon.Hash, logger *log.Logger) (*ethereum.Receipt, error) {
	logger.WithField("hash", txHash.String()).Debug("fetching receipt")
	receipt, err := uc.ethClient.TransactionReceipt(ctx, chainURL, txHash)
	if err != nil {
		logger.WithError(err).Error("failed to fetch receipt")
		return nil, err
	}

	return receipt.
		SetBlockHash(ethcommon.HexToHash(receipt.GetBlockHash())).
		SetBlockNumber(receipt.GetBlockNumber()).
		SetTxIndex(receipt.TxIndex), nil
}

func (uc *updatedJobStatusUseCase) fetchPrivateReceipt(ctx context.Context, chainURL string, txHash ethcommon.Hash, logger *log.Logger) (*ethereum.Receipt, error) {
	logger.WithField("hash", txHash.String()).Debug("fetching private receipt")

	receipt, err := uc.ethClient.PrivateTransactionReceipt(ctx, chainURL, txHash)

	// We exit ONLY when we failed to fetch the marking tx receipt, otherwise
	// error is being appended to the envelope
	if err != nil && receipt == nil {
		errMsg := "failed to fetch private receipt"
		logger.WithError(err).Error(errMsg)
		return nil, errors.DependencyFailureError(errMsg)
	} else if receipt == nil {
		errMsg := "fetched an empty receipt"
		logger.WithError(err).Error(errMsg)
		return nil, errors.InvalidParameterError(errMsg)
	}

	// Bind the hybrid receipt to the envelope
	return receipt.
		SetBlockHash(ethcommon.HexToHash(receipt.GetBlockHash())).
		SetBlockNumber(receipt.GetBlockNumber()).
		SetTxHash(txHash).
		SetTxIndex(receipt.TxIndex), nil
}
