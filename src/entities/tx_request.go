package entities

import (
	"time"
)

type TxRequest struct {
	IdempotencyKey string
	ChainName      string
	Hash           string
	Schedule       *Schedule
	Params         *TxRequestParams
	Labels         map[string]string
	InternalData   *InternalData
	CreatedAt      time.Time
}

type TxRequestParams struct {
	*ETHTransaction
	ContractTag     string               `json:"contractTag,omitempty"`
	ContractName    string               `json:"contractName,omitempty"`
	MethodSignature string               `json:"methodSignature,omitempty"`
	Args            []interface{}        `json:"args,omitempty"`
	Protocol        PrivateTxManagerType `json:"protocol,omitempty" example:"Tessera"`
}
