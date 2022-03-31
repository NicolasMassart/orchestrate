package entities

import (
	"math/big"
	"time"

	"github.com/consensys/orchestrate/pkg/types/ethereum"
	quorumtypes "github.com/consensys/quorum/core/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type TransactionType string

var (
	LegacyTxType     TransactionType = "legacy"
	DynamicFeeTxType TransactionType = "dynamic_fee"
)

func (tt *TransactionType) String() string {
	return string(*tt)
}

type PrivacyFlag int

const (
	PrivacyFlagSP  PrivacyFlag = iota
	PrivacyFlagPP  PrivacyFlag = 1
	PrivacyFlagMPP PrivacyFlag = 2
	PrivacyFlagPSV PrivacyFlag = 3
)

type ETHTransaction struct {
	Hash            *ethcommon.Hash
	From            *ethcommon.Address
	To              *ethcommon.Address
	Nonce           *uint64
	Value           *hexutil.Big
	Gas             *uint64
	GasPrice        *hexutil.Big
	GasFeeCap       *hexutil.Big
	GasTipCap       *hexutil.Big
	AccessList      types.AccessList
	TransactionType TransactionType
	Data            hexutil.Bytes
	Raw             hexutil.Bytes
	PrivateFrom     string
	PrivateFor      []string
	MandatoryFor    []string
	PrivacyGroupID  string
	PrivacyFlag     PrivacyFlag
	EnclaveKey      hexutil.Bytes
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (tx *ETHTransaction) ToETHTransaction(chainID *big.Int) *types.Transaction {
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
	case LegacyTxType:
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

func (tx *ETHTransaction) ToQuorumTransaction() *quorumtypes.Transaction {
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

func ConvertFromAccessList(accessList types.AccessList) []*ethereum.AccessTuple {
	result := []*ethereum.AccessTuple{}
	for _, t := range accessList {
		tupl := &ethereum.AccessTuple{
			Address:     t.Address.Hex(),
			StorageKeys: []string{},
		}

		for _, k := range t.StorageKeys {
			tupl.StorageKeys = append(tupl.StorageKeys, k.Hex())
		}

		result = append(result, tupl)
	}

	return result
}

func ConvertToAccessList(accessList []*ethereum.AccessTuple) types.AccessList {
	result := types.AccessList{}
	for _, item := range accessList {
		storageKeys := []ethcommon.Hash{}
		for _, sk := range item.StorageKeys {
			storageKeys = append(storageKeys, ethcommon.HexToHash(sk))
		}

		result = append(result, types.AccessTuple{
			Address:     ethcommon.HexToAddress(item.Address),
			StorageKeys: storageKeys,
		})
	}

	return result
}
