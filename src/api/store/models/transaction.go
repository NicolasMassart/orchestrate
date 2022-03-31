package models

import (
	"time"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"

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
	AccessList     types.AccessList
	PrivateFrom    string
	PrivateFor     []string `pg:",array"`
	MandatoryFor   []string `pg:",array"`
	PrivacyGroupID string
	PrivacyFlag    int
	EnclaveKey     string    `pg:"alias:enclave_key"`
	CreatedAt      time.Time `pg:"default:now()"`
	UpdatedAt      time.Time `pg:"default:now()"`
}

func NewTransaction(tx *entities.ETHTransaction) *Transaction {
	return &Transaction{
		Hash:           utils.StringerToString(tx.Hash),
		Sender:         utils.StringerToString(tx.From),
		Recipient:      utils.StringerToString(tx.To),
		Nonce:          utils.ValueToString(tx.Nonce),
		Value:          utils.HexToBigIntString(tx.Value),
		GasPrice:       utils.HexToBigIntString(tx.GasPrice),
		GasFeeCap:      utils.HexToBigIntString(tx.GasFeeCap),
		GasTipCap:      utils.HexToBigIntString(tx.GasTipCap),
		Gas:            utils.ValueToString(tx.Gas),
		Data:           utils.StringerToString(tx.Data),
		Raw:            utils.StringerToString(tx.Raw),
		TxType:         string(tx.TransactionType),
		AccessList:     tx.AccessList,
		PrivateFrom:    tx.PrivateFrom,
		PrivateFor:     tx.PrivateFor,
		MandatoryFor:   tx.MandatoryFor,
		PrivacyGroupID: tx.PrivacyGroupID,
		PrivacyFlag:    int(tx.PrivacyFlag),
		EnclaveKey:     utils.StringerToString(tx.EnclaveKey),
		CreatedAt:      tx.CreatedAt,
		UpdatedAt:      tx.UpdatedAt,
	}
}

func (tx *Transaction) ToEntity() *entities.ETHTransaction {
	return &entities.ETHTransaction{
		Hash:            utils.StringToETHHash(tx.Hash),
		From:            utils.ToEthAddr(tx.Sender),
		To:              utils.ToEthAddr(tx.Recipient),
		Nonce:           utils.StringToUint64(tx.Nonce),
		Value:           utils.StringBigIntToHex(tx.Value),
		GasPrice:        utils.StringBigIntToHex(tx.GasPrice),
		Gas:             utils.StringToUint64(tx.Gas),
		GasTipCap:       utils.StringBigIntToHex(tx.GasTipCap),
		GasFeeCap:       utils.StringBigIntToHex(tx.GasFeeCap),
		Data:            utils.StringToHexBytes(tx.Data),
		TransactionType: entities.TransactionType(tx.TxType),
		AccessList:      tx.AccessList,
		PrivateFrom:     tx.PrivateFrom,
		PrivateFor:      tx.PrivateFor,
		MandatoryFor:    tx.MandatoryFor,
		PrivacyGroupID:  tx.PrivacyGroupID,
		PrivacyFlag:     entities.PrivacyFlag(tx.PrivacyFlag),
		EnclaveKey:      utils.StringToHexBytes(tx.EnclaveKey),
		Raw:             utils.StringToHexBytes(tx.Raw),
		CreatedAt:       tx.CreatedAt,
		UpdatedAt:       tx.UpdatedAt,
	}
}
