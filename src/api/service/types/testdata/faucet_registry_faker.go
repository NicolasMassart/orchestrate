package testdata

import (
	"math/big"

	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/quorum-key-manager/pkg/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gofrs/uuid"
)

func FakeRegisterFaucetRequest() *api.RegisterFaucetRequest {
	return &api.RegisterFaucetRequest{
		Name:            "faucet-chain-" + common.RandString(5),
		ChainRule:       uuid.Must(uuid.NewV4()).String(),
		CreditorAccount: *testdata.FakeAddress(),
		MaxBalance:      hexutil.Big(*big.NewInt(6000000)),
		Amount:          hexutil.Big(*big.NewInt(100)),
		Cooldown:        "10s",
	}
}

func FakeUpdateFaucetRequest() *api.UpdateFaucetRequest {
	return &api.UpdateFaucetRequest{
		Name:            "faucet-chain-"+ common.RandString(5),
		ChainRule:       uuid.Must(uuid.NewV4()).String(),
		CreditorAccount: testdata.FakeAddress(),
		MaxBalance:      hexutil.Big(*big.NewInt(6000000)),
		Amount:          hexutil.Big(*big.NewInt(100)),
		Cooldown:        "10s",
	}
}
