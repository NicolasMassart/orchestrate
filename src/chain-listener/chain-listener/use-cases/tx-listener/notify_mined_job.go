package txlistener

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	pkgbackoff "github.com/consensys/orchestrate/pkg/backoff"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/ethereum/abi"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/types"
	usecases "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/chain-listener/store"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	ethAbi "github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const notifyMinedJobTxComponent = "chain-listener.use-case.tx-listener.notify-mined-job"

type notifyMinedJobUC struct {
	client                   orchestrateclient.OrchestrateClient
	ethClient                ethclient.MultiClient
	sendNotification         usecases.SendNotification
	registerDeployedContract usecases.RegisterDeployedContract
	chainState               store.Chain
	logger                   *log.Logger
}

func NotifyMinedJobUseCase(client orchestrateclient.OrchestrateClient,
	ethClient ethclient.MultiClient,
	sendNotification usecases.SendNotification,
	registerDeployedContract usecases.RegisterDeployedContract,
	chainState store.Chain,
	logger *log.Logger,
) usecases.NotifyMinedJob {
	return &notifyMinedJobUC{
		client:                   client,
		ethClient:                ethClient,
		sendNotification:         sendNotification,
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

	if err := uc.updateJobStatus(ctx, job, job.Receipt, logger); err != nil {
		return err
	}

	if err := uc.sendNotification.Execute(ctx, job); err != nil {
		return err
	}

	return nil
}

func (uc *notifyMinedJobUC) updateJobStatus(ctx context.Context, job *entities.Job, receipt *ethereum.Receipt, logger *log.Logger) error {
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

	err = uc.decodeLogs(ctx, job.ChainUUID, receipt)
	if err != nil {
		return receipt, err
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

func (uc *notifyMinedJobUC) decodeLogs(ctx context.Context, chainUUID string, receipt *ethereum.Receipt) error {
	uc.logger.Debug("decoding receipt logs...")
	var contractAddress *ethcommon.Address

	chain, err := uc.chainState.Get(ctx, chainUUID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			uc.logger.Warn(err)
			return nil
		}
		errMsg := "failed to get chain for decoding logs"
		uc.logger.WithError(err).Error(errMsg)
		return errors.DependencyFailureError(errMsg)
	}

	for _, l := range receipt.GetLogs() {
		if len(l.GetTopics()) == 0 {
			// This scenario is not supposed to happen
			return errors.DependencyFailureError("invalid receipt (no topics in log)")
		}

		logger := uc.logger.WithContext(ctx).WithField("sig_hash", utils.ShortString(l.Topics[0], 5)).
			WithField("address", l.GetAddress()).WithField("indexed", uint32(len(l.Topics)-1))

		logger.Debug("decoding receipt logs")
		sigHash := hexutil.MustDecode(l.Topics[0])
		eventResp, err := uc.client.GetContractEvents(
			ctx,
			l.GetAddress(),
			chain.ChainID.String(),
			&types.GetContractEventsRequest{
				SigHash:           sigHash,
				IndexedInputCount: uint32(len(l.Topics) - 1),
			},
		)

		if err != nil {
			if errors.IsNotFoundError(err) {
				continue
			}

			logger.WithError(err).Error("failed to decode receipt logs")
			return err
		}

		if eventResp.Event == "" && len(eventResp.DefaultEvents) == 0 {
			logger.WithError(err).Trace("could not retrieve event ABI")
			continue
		}

		var mapping map[string]string
		event := &ethAbi.Event{}

		if eventResp.Event != "" {
			err = json.Unmarshal([]byte(eventResp.Event), event)
			if err != nil {
				logger.WithError(err).
					Warnf("could not unmarshal event ABI provided by the Contract Registry, txHash: %s sigHash: %s, ", l.GetTxHash(), l.GetTopics()[0])
				continue
			}
			mapping, err = abi.Decode(event, l)
		} else {
			for _, potentialEvent := range eventResp.DefaultEvents {
				// Try to unmarshal
				err = json.Unmarshal([]byte(potentialEvent), event)
				if err != nil {
					// If it fails to unmarshal, try the next potential event
					logger.WithError(err).Tracef("could not unmarshal potential event ABI, txHash: %s sigHash: %s, ", l.GetTxHash(), l.GetTopics()[0])
					continue
				}

				// Try to decode
				mapping, err = abi.Decode(event, l)
				if err == nil {
					// As the decoding is successful, stop looping
					break
				}
			}
		}

		if err != nil {
			// As all potentialEvents fail to unmarshal, go to the next log
			logger.WithError(err).Tracef("could not unmarshal potential event ABI, txHash: %s sigHash: %s, ", l.GetTxHash(), l.GetTopics()[0])
			continue
		}

		if contractAddress == nil {
			addr := ethcommon.HexToAddress(l.GetAddress())
			contractAddress = &addr
		}

		// Set decoded data on log
		l.DecodedData = mapping
		l.Event = getAbi(event)

		logger.WithField("receipt_log", fmt.Sprintf("%v", mapping)).
			Debug("log decoded")
	}

	return nil
}

// GetAbi creates a string ABI (format EventName(argType1, argType2)) from an event
func getAbi(e *ethAbi.Event) string {
	inputs := make([]string, len(e.Inputs))
	for i := range e.Inputs {
		inputs[i] = fmt.Sprintf("%v", e.Inputs[i].Type)
	}
	return fmt.Sprintf("%v(%v)", e.Name, strings.Join(inputs, ","))
}
