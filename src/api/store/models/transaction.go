package models

import (
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

type Transaction struct {
	tableName struct{} `pg:"transactions"` // nolint:unused,structcheck // reason

	ID             int
	UUID           string
	Hash           string
	Sender         string
	Recipient      string
	Nonce          string
	Value          string
	GasPrice       string
	GasFeeCap      string
	GasTipCap      string
	Gas            string
	Data           string
	Raw            string
	TxType         string
	AccessList     types.AccessList `pg:",json"`
	PrivateFrom    string
	PrivateFor     []string `pg:",array"`
	MandatoryFor   []string `pg:",array"`
	PrivacyGroupID string
	PrivacyFlag    int
	EnclaveKey     string    `pg:"alias:enclave_key"`
	CreatedAt      time.Time `pg:"default:now()"`
	UpdatedAt      time.Time `pg:"default:now()"`
}
