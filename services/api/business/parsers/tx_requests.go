package parsers

import (
	"fmt"

	"github.com/ConsenSys/orchestrate/pkg/types/entities"
	"github.com/ConsenSys/orchestrate/services/api/store/models"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func NewTxRequestModelFromEntities(txRequest *entities.TxRequest, requestHash string, scheduleID int) *models.TransactionRequest {
	return &models.TransactionRequest{
		IdempotencyKey: txRequest.IdempotencyKey,
		ChainName:      txRequest.ChainName,
		RequestHash:    requestHash,
		Params:         txRequest.Params,
		ScheduleID:     &scheduleID,
		CreatedAt:      txRequest.CreatedAt,
	}
}

func NewJobEntitiesFromTxRequest(txRequest *entities.TxRequest, chainUUID, txData string) ([]*entities.Job, error) {
	var jobs []*entities.Job
	switch {
	case txRequest.Params.Protocol == entities.OrionChainType:
		privTxJob := newJobEntityFromTxRequest(txRequest, newEthTransactionFromParams(txRequest.Params, txData), entities.OrionEEATransaction, chainUUID)
		markingTxJob := newJobEntityFromTxRequest(txRequest, &entities.ETHTransaction{}, entities.OrionMarkingTransaction, chainUUID)
		markingTxJob.InternalData.OneTimeKey = true
		jobs = append(jobs, privTxJob, markingTxJob)
	case txRequest.Params.Protocol == entities.TesseraChainType:
		privTxJob := newJobEntityFromTxRequest(txRequest, newEthTransactionFromParams(txRequest.Params, txData),
			entities.TesseraPrivateTransaction, chainUUID)
		markingTxJob := newJobEntityFromTxRequest(txRequest, &entities.ETHTransaction{From: txRequest.Params.From,
			PrivateFor: txRequest.Params.PrivateFor}, entities.TesseraMarkingTransaction, chainUUID)
		jobs = append(jobs, privTxJob, markingTxJob)
	case txRequest.Params.Raw != "":
		rawTx, err := newTransactionFromRaw(txRequest.Params.Raw)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, newJobEntityFromTxRequest(txRequest, rawTx, entities.EthereumRawTransaction, chainUUID))
	default:
		jobs = append(jobs, newJobEntityFromTxRequest(txRequest, newEthTransactionFromParams(txRequest.Params, txData),
			entities.EthereumTransaction, chainUUID))
	}

	return jobs, nil
}

func newEthTransactionFromParams(params *entities.ETHTransactionParams, txData string) *entities.ETHTransaction {
	return &entities.ETHTransaction{
		From:           params.From,
		To:             params.To,
		Nonce:          params.Nonce,
		Value:          params.Value,
		GasPrice:       params.GasPrice,
		Gas:            params.Gas,
		Raw:            params.Raw,
		Data:           txData,
		PrivateFrom:    params.PrivateFrom,
		PrivateFor:     params.PrivateFor,
		PrivacyGroupID: params.PrivacyGroupID,
	}
}

func newJobEntityFromTxRequest(txRequest *entities.TxRequest, ethTx *entities.ETHTransaction, jobType entities.JobType, chainUUID string) *entities.Job {
	internalData := *txRequest.InternalData
	return &entities.Job{
		ScheduleUUID: txRequest.Schedule.UUID,
		ChainUUID:    chainUUID,
		Type:         jobType,
		Labels:       txRequest.Labels,
		InternalData: &internalData,
		Transaction:  ethTx,
		TenantID:     txRequest.Schedule.TenantID,
	}
}

func newTransactionFromRaw(raw string) (*entities.ETHTransaction, error) {
	tx := &types.Transaction{}

	rawb, err := hexutil.Decode(raw)
	if err != nil {
		return nil, err
	}

	err = rlp.DecodeBytes(rawb, &tx)
	if err != nil {
		return nil, err
	}

	msg, err := tx.AsMessage(types.NewEIP155Signer(tx.ChainId()))
	if err != nil {
		return nil, err
	}

	jobTx := &entities.ETHTransaction{
		From:     msg.From().String(),
		Data:     hexutil.Encode(tx.Data()),
		Gas:      fmt.Sprintf("%d", tx.Gas()),
		GasPrice: fmt.Sprintf("%d", tx.GasPrice()),
		Value:    tx.Value().String(),
		Nonce:    fmt.Sprintf("%d", tx.Nonce()),
		Hash:     tx.Hash().String(),
		Raw:      raw,
	}

	// If not contract creation
	if tx.To() != nil {
		jobTx.To = tx.To().String()
	}

	return jobTx, nil
}
