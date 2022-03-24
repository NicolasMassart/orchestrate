package entities

import (
	"time"

	"github.com/consensys/orchestrate/pkg/types/ethereum"
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
