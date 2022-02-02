package types

import (
	"time"
)

type ETHTransactionResponse struct {
	Hash            string                `json:"hash,omitempty"`
	From            string                `json:"from,omitempty"`
	To              string                `json:"to,omitempty"`
	Nonce           *uint64               `json:"nonce,omitempty"`
	Value           string                `json:"value,omitempty"`
	Gas             *uint64               `json:"gas,omitempty"`
	GasPrice        string                `json:"gasPrice,omitempty"`
	GasFeeCap       string                `json:"maxFeePerGas,omitempty"`
	GasTipCap       string                `json:"maxPriorityFeePerGas,omitempty"`
	AccessList      []AccessTupleResponse `json:"accessList,omitempty"`
	TransactionType string                `json:"transactionType,omitempty"`
	Data            string                `json:"data,omitempty"`
	Raw             string                `json:"raw,omitempty"`
	PrivateFrom     string                `json:"privateFrom,omitempty"`
	PrivateFor      []string              `json:"privateFor,omitempty"`
	MandatoryFor    []string              `json:"mandatoryFor,omitempty"`
	PrivacyGroupID  string                `json:"privacyGroupId,omitempty"`
	PrivacyFlag     int                   `json:"privacyFlag,omitempty"`
	EnclaveKey      string                `json:"enclaveKey,omitempty"`
	CreatedAt       time.Time             `json:"createdAt,omitempty"`
	UpdatedAt       time.Time             `json:"updatedAt,omitempty"`
}

type AccessTupleResponse struct {
	Address     string   `json:"address"       `
	StorageKeys []string `json:"storageKeys"   `
}
