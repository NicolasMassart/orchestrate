package testutils

import (
	"time"

	"github.com/ConsenSys/orchestrate/pkg/types/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/ConsenSys/orchestrate/pkg/utils"
)

func ParseIArray(args ...interface{}) (ret []interface{}) {
	ret = make([]interface{}, len(args))
	copy(ret, args)
	return
}

func FakeETHTransaction() *entities.ETHTransaction {
	return &entities.ETHTransaction{
		From:        FromAddress,
		To:          "0x4FED1fC4144c223aE3C1553be203cDFcbD38C581",
		Nonce:       "1",
		Value:       "50000",
		GasPrice:    "10000",
		Gas:         "21000",
		Data:        "0x",
		Raw:         "0x",
		PrivateFrom: "A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=",
		PrivateFor:  []string{"A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=", "B1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func FakeETHTransactionParams() *entities.ETHTransactionParams {
	return &entities.ETHTransactionParams{
		From:            "0x7357589f8e367c2C31F51242fB77B350A11830F3",
		To:              "0x7357589f8e367c2C31F51242fB77B350A11830F2",
		Value:           "1",
		GasPrice:        "0",
		Gas:             "0",
		MethodSignature: "method(string,string)",
		Args:            ParseIArray("val1", "val2"),
		ContractName:    "ContractName",
		ContractTag:     "ContractTag",
		Nonce:           "1",
	}
}

func FakePrivateETHTransactionParams() *entities.PrivateETHTransactionParams {
	return &entities.PrivateETHTransactionParams{
		PrivateFrom:   "A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=",
		PrivateFor:    []string{"A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="},
		PrivateTxType: entities.PrivateTxTypeRestricted,
	}
}

func FakeETHAccount() *entities.ETHAccount {
	return &entities.ETHAccount{
		Namespace:           "_",
		Address:             ethcommon.HexToAddress(utils.RandHexString(12)).String(),
		PublicKey:           ethcommon.HexToHash(utils.RandHexString(12)).String(),
		CompressedPublicKey: ethcommon.HexToHash(utils.RandHexString(12)).String(),
	}
}

func FakeZKSAccount() *entities.ZKSAccount {
	return &entities.ZKSAccount{
		Namespace:        "_",
		PublicKey:        ethcommon.HexToHash(utils.RandHexString(12)).String(),
		Curve:            entities.ZKSCurveBN254,
		SigningAlgorithm: entities.ZKSAlgorithmEDDSA,
	}
}

func FakeTesseraTransactionParams() *entities.ETHTransactionParams {
	tx := FakeETHTransactionParams()
	tx.PrivateFrom = "ROAZBWtSacxXQrOe3FGAqJDyJjFePR5ce4TSIzmJ0Bc="
	tx.PrivateFor = []string{"ROAZBWtSacxXQrOe3FGAqJDyJjFePR5ce4TSIzmJ0Bd="}
	tx.Protocol = entities.TesseraChainType

	return tx
}

func FakeOrionTransactionParams() *entities.ETHTransactionParams {
	tx := FakeETHTransactionParams()
	tx.PrivateFrom = "ROAZBWtSacxXQrOe3FGAqJDyJjFePR5ce4TSIzmJ0Be="
	tx.PrivacyGroupID = "ROAZBWtSacxXQrOe3FGAqJDyJjFePR5ce4TSIzmJ0Bf="
	tx.Protocol = entities.OrionChainType

	return tx
}

func FakeRawTransactionParams() *entities.ETHTransactionParams {
	return &entities.ETHTransactionParams{
		Raw: "0xABCDE012312312",
	}
}

func FakeTransferTransactionParams() *entities.ETHTransactionParams {
	return &entities.ETHTransactionParams{
		From:  "0x7357589f8e367c2C31F51242fB77B350A11830FA",
		To:    "0x7357589f8e367c2C31F51242fB77B350A11830FB",
		Value: "10000000000",
	}
}

func FakeAddress() ethcommon.Address {
	return ethcommon.HexToAddress(utils.RandHexString(20))
}

func FakeHash() ethcommon.Hash {
	return ethcommon.HexToHash(utils.RandHexString(40))
}
