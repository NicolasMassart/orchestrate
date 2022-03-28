package txlistener

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	pkgbackoff "github.com/consensys/orchestrate/pkg/backoff"
	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const notifyMinedJobTxComponent = "tx-listener.use-case.tx-listener.notify-mined-job"

type notifyMinedJobUC struct {
	client                   orchestrateclient.OrchestrateClient
	ethClient                ethclient.MultiClient
	registerDeployedContract usecases.RegisterDeployedContract
	chainState               store.Chain
	logger                   *log.Logger
}

func NotifyMinedJobUseCase(client orchestrateclient.OrchestrateClient,
	ethClient ethclient.MultiClient,
	registerDeployedContract usecases.RegisterDeployedContract,
	chainState store.Chain,
	logger *log.Logger,
) usecases.NotifyMinedJob {
	return &notifyMinedJobUC{
		client:                   client,
		ethClient:                ethClient,
		registerDeployedContract: registerDeployedContract,
		chainState:               chainState,
		logger:                   logger.SetComponent(notifyMinedJobTxComponent),
	}
}

func (uc *notifyMinedJobUC) Execute(ctx context.Context, job *entities.Job) error {
	logger := uc.logger.WithField("job", job.UUID).WithField("tx_hash", job.Transaction.Hash)
	logger.Debug("updating job to mined")

	// There is a racing issue between tx included in the block and receipt being available
	err := backoff.RetryNotify(
		func() error {
			var err error
			job.Receipt, err = uc.getTxReceipt(ctx, job, logger)
			return err
		},
		pkgbackoff.ConstantBackOffWithMaxRetries(time.Second, 3),
		func(err error, d time.Duration) {
			logger.WithError(err).Warnf("error fetching receipt, restarting in %v...", d)
		},
	)
	if err != nil {
		errMsg := "failed to fetch receipt"
		logger.WithError(err).Error(errMsg)
		return errors.DependencyFailureError(errMsg)
	}

	// If contract has been deployed
	if job.Receipt.ContractAddress != "" && job.Receipt.ContractAddress != utils.ZeroAddressString {
		if err := uc.registerDeployedContract.Execute(ctx, job); err != nil {
			return err
		}
	}

	if err := uc.updateJobStatus(ctx, job, logger); err != nil {
		return err
	}

	return nil
}

func (uc *notifyMinedJobUC) updateJobStatus(ctx context.Context, job *entities.Job, logger *log.Logger) error {
	updateTxReq := &types.UpdateJobRequest{
		Status:  entities.StatusMined,
		Message: fmt.Sprintf("transaction mined in block %v", job.Receipt.BlockNumber),
		Receipt: job.Receipt,
	}

	if job.Transaction.TransactionType == entities.DynamicFeeTxType {
		effectiveGas, _ := hexutil.DecodeBig(job.Receipt.EffectiveGasPrice)
		updateTxReq.Transaction = &types.ETHTransactionRequest{
			GasPrice: (*hexutil.Big)(effectiveGas),
		}
	}

	_, err := uc.client.UpdateJob(ctx, job.UUID, updateTxReq)
	if err != nil {
		errMsg := "failed to update job to MINED"
		logger.WithError(err).Error(errMsg)
		return errors.DependencyFailureError(errMsg)
	}

	logger.Info("job was updated to mined successfully")
	return nil
}

func (uc *notifyMinedJobUC) getTxReceipt(ctx context.Context, job *entities.Job, logger *log.Logger) (*ethereum.Receipt, error) {
	logger.Debug("fetching transaction receipt")
	chainURL := uc.client.ChainProxyURL(job.ChainUUID)

	var err error
	var receipt *ethereum.Receipt
	switch job.Type {
	case entities.EEAMarkingTransaction:
		receipt, err = uc.fetchPrivateReceipt(ctx, chainURL, *job.Transaction.Hash)
	default:
		receipt, err = uc.fetchReceipt(ctx, chainURL, *job.Transaction.Hash)
	}
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (uc *notifyMinedJobUC) fetchReceipt(ctx context.Context, chainURL string, txHash ethcommon.Hash) (*ethereum.Receipt, error) {
	receipt, err := uc.ethClient.TransactionReceipt(ctx, chainURL, txHash)
	if err != nil {
		return nil, err
	}

	return receipt.
		SetBlockHash(ethcommon.HexToHash(receipt.GetBlockHash())).
		SetBlockNumber(receipt.GetBlockNumber()).
		SetTxIndex(receipt.TxIndex), nil
}

func (uc *notifyMinedJobUC) fetchPrivateReceipt(ctx context.Context, chainURL string, txHash ethcommon.Hash) (*ethereum.Receipt, error) {
	receipt, err := uc.ethClient.PrivateTransactionReceipt(ctx, chainURL, txHash)

	// We exit ONLY when we failed to fetch the marking tx receipt, otherwise
	// error is being appended to the envelope
	if err != nil && receipt == nil {
		return nil, err
	} else if receipt == nil {
		return nil, errors.InvalidParameterError("fetched an empty private receipt")
	}

	// Bind the hybrid receipt to the envelope
	return receipt.
		SetBlockHash(ethcommon.HexToHash(receipt.GetBlockHash())).
		SetBlockNumber(receipt.GetBlockNumber()).
		SetTxHash(txHash).
		SetTxIndex(receipt.TxIndex), nil
}
