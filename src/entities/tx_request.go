package entities

import (
	"time"
)

type TxRequest struct {
	IdempotencyKey string
	ChainName      string
	Hash           string
	Schedule       *Schedule
	Params         *ETHTransactionParams
	Labels         map[string]string
	InternalData   *InternalData
	CreatedAt      time.Time
}
