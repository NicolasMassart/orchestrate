package types

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type RawTransactionRequest struct {
	ChainName string               `json:"chain" validate:"required" example:"myChain"` // Name of the chain on which to send the transaction.
	Labels    map[string]string    `json:"labels,omitempty"`                            // List of custom labels.
	Params    RawTransactionParams `json:"params" validate:"required"`
}
type RawTransactionParams struct {
	Raw         hexutil.Bytes       `json:"raw" validate:"required" example:"0xfe378324abcde723" swaggertype:"string"` // Transaction raw data.
	RetryPolicy IntervalRetryParams `json:"retryPolicy,omitempty"`
}
