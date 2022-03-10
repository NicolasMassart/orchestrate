package testdata

import (
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/quorum/common/hexutil"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

func FakeAccount() *entities.Account {
	return &entities.Account{
		Alias:               "MyAccount",
		TenantID:            multitenancy.DefaultTenant,
		StoreID:             "qkm-store-id",
		OwnerID:             "owner",
		Attributes:          make(map[string]string),
		Address:             ethcommon.HexToAddress("0x5Cc634233E4a454d47aACd9fC68801482Fb02610"),
		PublicKey:           hexutil.MustDecode("0x" + utils.RandHexString(12)),
		CompressedPublicKey: hexutil.MustDecode("0x" + utils.RandHexString(24)),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}
