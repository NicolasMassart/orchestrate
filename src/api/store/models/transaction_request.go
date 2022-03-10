package models

import (
	"time"

	"github.com/consensys/orchestrate/pkg/utils"

	"github.com/consensys/orchestrate/src/entities"
)

type TransactionRequest struct {
	tableName struct{} `pg:"transaction_requests"` // nolint:unused,structcheck // reason

	ID             int
	IdempotencyKey string
	ChainName      string
	ScheduleID     *int
	Schedule       *Schedule `pg:"rel:has-one"`
	RequestHash    string
	Params         *entities.TxRequestParams
	CreatedAt      time.Time `pg:"default:now()"`
}

func NewTxRequest(txRequest *entities.TxRequest, requestHash string) *TransactionRequest {
	return &TransactionRequest{
		IdempotencyKey: txRequest.IdempotencyKey,
		ChainName:      txRequest.ChainName,
		RequestHash:    requestHash,
		Params:         txRequest.Params,
		CreatedAt:      txRequest.CreatedAt,
	}
}

func NewTxRequests(txRequests []*TransactionRequest) []*entities.TxRequest {
	res := []*entities.TxRequest{}
	for _, req := range txRequests {
		res = append(res, req.ToEntity())
	}

	return res
}

func (txRequest *TransactionRequest) ToEntity() *entities.TxRequest {
	res := &entities.TxRequest{
		IdempotencyKey: txRequest.IdempotencyKey,
		ChainName:      txRequest.ChainName,
		CreatedAt:      txRequest.CreatedAt,
		Hash:           txRequest.RequestHash,
		Params:         txRequest.Params,
	}

	if txRequest.Params != nil {
		res.Params = &entities.TxRequestParams{}
		_ = utils.CopyInterface(txRequest.Params, res.Params)
	}

	if txRequest.Schedule != nil {
		res.Schedule = txRequest.Schedule.ToEntity()
	}

	return res
}
