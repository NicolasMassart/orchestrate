package signer

import (
	"encoding/base64"
	"math/big"

	"github.com/consensys/orchestrate/pkg/encoding/rlp"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
	quorumtypes "github.com/consensys/quorum/core/types"
	"github.com/ethereum/go-ethereum/core/types"
)

func ethTransactionToTransaction(tx *entities.ETHTransaction, chainID *big.Int) *types.Transaction {
	var txData types.TxData

	var value *big.Int
	if tx.Value != nil {
		value = tx.Value.ToInt()
	}

	var gasPrice *big.Int
	if tx.GasPrice != nil {
		gasPrice = tx.GasPrice.ToInt()
	}
	var nonce uint64
	if tx.Nonce != nil {
		nonce = *tx.Nonce
	}
	var gasLimit uint64
	if tx.Gas != nil {
		gasLimit = *tx.Gas
	}

	switch tx.TransactionType {
	case entities.LegacyTxType:
		txData = &types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gasLimit,
			To:       tx.To,
			Value:    value,
			Data:     tx.Data,
		}
	default:
		var gasTipCap *big.Int
		if tx.GasTipCap != nil {
			gasTipCap = tx.GasTipCap.ToInt()
		}

		var gasFeeCap *big.Int
		if tx.GasFeeCap != nil {
			gasFeeCap = tx.GasFeeCap.ToInt()
		}

		txData = &types.DynamicFeeTx{
			ChainID:    chainID,
			Nonce:      nonce,
			GasTipCap:  gasTipCap,
			GasFeeCap:  gasFeeCap,
			Gas:        gasLimit,
			To:         tx.To,
			Value:      value,
			Data:       tx.Data,
			AccessList: tx.AccessList,
		}
	}

	return types.NewTx(txData)
}

func ethTransactionToQuorumTransaction(tx *entities.ETHTransaction) *quorumtypes.Transaction {
	var value *big.Int
	if tx.Value != nil {
		value = tx.Value.ToInt()
	}

	var gasPrice *big.Int
	if tx.GasPrice != nil {
		gasPrice = tx.GasPrice.ToInt()
	}
	var nonce uint64
	if tx.Nonce != nil {
		nonce = *tx.Nonce
	}
	var gasLimit uint64
	if tx.Gas != nil {
		gasLimit = *tx.Gas
	}

	if tx.To == nil {
		return quorumtypes.NewContractCreation(nonce, value, gasLimit, gasPrice, tx.Data)
	}

	return quorumtypes.NewTransaction(nonce, *tx.To, value, gasLimit, gasPrice, tx.Data)
}

func eeaTransactionPayload(tx *types.Transaction, privateArgs *privateETHTransactionParams, chainID *big.Int) ([]byte, error) {
	privateFromEncoded, err := getEncodedPrivateFrom(privateArgs.PrivateFrom)
	if err != nil {
		return nil, err
	}

	privateRecipientEncoded, err := getEncodedPrivateRecipient(privateArgs.PrivacyGroupID, privateArgs.PrivateFor)
	if err != nil {
		return nil, err
	}

	hash, err := rlp.Hash([]interface{}{
		tx.Nonce(),
		tx.GasPrice(),
		tx.Gas(),
		tx.To(),
		tx.Value(),
		tx.Data(),
		chainID,
		uint(0),
		uint(0),
		privateFromEncoded,
		privateRecipientEncoded,
		privateArgs.PrivateTxType,
	})
	if err != nil {
		return nil, errors.CryptoOperationError("failed to hash eea transaction").AppendReason(err.Error())
	}

	return hash.Bytes(), nil
}

func getEncodedPrivateFrom(privateFrom string) ([]byte, error) {
	privateFromEncoded, err := base64.StdEncoding.DecodeString(privateFrom)
	if err != nil {
		return nil, errors.EncodingError("invalid base64 value for 'privateFrom'").AppendReason(err.Error())
	}

	return privateFromEncoded, nil
}

func getEncodedPrivateRecipient(privacyGroupID string, privateFor []string) (interface{}, error) {
	var privateRecipientEncoded interface{}
	var err error
	if privacyGroupID != "" {
		privateRecipientEncoded, err = base64.StdEncoding.DecodeString(privacyGroupID)
		if err != nil {
			return nil, errors.EncodingError("invalid base64 value for 'privacyGroupId'").AppendReason(err.Error())
		}
	} else {
		var privateForByteSlice [][]byte
		for _, v := range privateFor {
			b, der := base64.StdEncoding.DecodeString(v)
			if der != nil {
				return nil, errors.EncodingError("invalid base64 value for 'privateFor'").AppendReason(der.Error())
			}
			privateForByteSlice = append(privateForByteSlice, b)
		}
		privateRecipientEncoded = privateForByteSlice
	}

	return privateRecipientEncoded, nil
}

func getQuorumPrivateTxSigner() quorumtypes.Signer {
	return quorumtypes.QuorumPrivateTxSigner{}
}
