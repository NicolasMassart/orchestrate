package testdata

import (
	"math/big"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func FakeETHTransactionReq() *types.ETHTransactionRequest {
	return &types.ETHTransactionRequest{
		From:        testdata.FakeAddress(),
		To:          testdata.FakeAddress(),
		Nonce:       utils.ToPtr(uint64(1)).(*uint64),
		Value:       utils.ToPtr(hexutil.Big(*big.NewInt(50000))).(*hexutil.Big),
		GasPrice:    utils.ToPtr(hexutil.Big(*big.NewInt(10000))).(*hexutil.Big),
		Gas:         utils.ToPtr(uint64(21000)).(*uint64),
		Data:        hexutil.MustDecode("0x0A"),
		Raw:         hexutil.MustDecode("0x0BCD"),
		PrivateFrom: "A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=",
		PrivateFor:  []string{"A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=", "B1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="},
	}
}
