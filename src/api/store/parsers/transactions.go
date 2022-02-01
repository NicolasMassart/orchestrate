package parsers

import (
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/ethereum/go-ethereum/core/types"
)

func NewTransactionModel(tx *entities.ETHTransaction) *models.Transaction {
	return &models.Transaction{
		UUID:           tx.UUID,
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

func NewTransactionEntity(tx *models.Transaction) *entities.ETHTransaction {
	accessList := types.AccessList{}
	_ = utils.CastInterfaceToObject(tx.AccessList, &accessList)

	return &entities.ETHTransaction{
		UUID:            tx.UUID,
		Hash:            utils.StringToEthHash(tx.Hash),
		From:            utils.ToEthAddr(tx.Sender),
		To:              utils.ToEthAddr(tx.Recipient),
		Nonce:           utils.StringToUint64(tx.Nonce),
		Value:           utils.BigIntStringToHex(tx.Value),
		GasPrice:        utils.BigIntStringToHex(tx.GasPrice),
		Gas:             utils.StringToUint64(tx.Gas),
		GasTipCap:       utils.BigIntStringToHex(tx.GasTipCap),
		GasFeeCap:       utils.BigIntStringToHex(tx.GasFeeCap),
		Data:            utils.StringToHexBytes(tx.Data),
		TransactionType: entities.TransactionType(tx.TxType),
		AccessList:      accessList,
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
