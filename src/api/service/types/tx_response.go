package types

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
)

type TransactionResponse struct {
	UUID           string                         `json:"uuid" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"` // UUID of the transaction.
	IdempotencyKey string                         `json:"idempotencyKey" example:"myIdempotencyKey"`           // Idempotency key of the transaction request.
	ChainName      string                         `json:"chain" example:"myChain"`                             // Chain on which the transaction was created.
	Params         *entities.ETHTransactionParams `json:"params"`
	Jobs           []*JobResponse                 `json:"jobs"`                                            // List of jobs in the transaction.
	CreatedAt      time.Time                      `json:"createdAt" example:"2020-07-09T12:35:42.115395Z"` // Date and time at which the transaction was created.
}
