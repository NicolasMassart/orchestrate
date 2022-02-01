package parsers

import (
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/quorum/common/hexutil"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

func NewAccountModel(account *entities.Account) *models.Account {
	return &models.Account{
		Alias:               account.Alias,
		Address:             account.Address.String(),
		PublicKey:           account.PublicKey.String(),
		CompressedPublicKey: account.CompressedPublicKey.String(),
		TenantID:            account.TenantID,
		OwnerID:             account.OwnerID,
		StoreID:             account.StoreID,
		Attributes:          account.Attributes,
		CreatedAt:           account.CreatedAt,
		UpdatedAt:           account.UpdatedAt,
	}
}

func NewAccountEntity(account *models.Account) *entities.Account {
	return &entities.Account{
		Alias:               account.Alias,
		Address:             ethcommon.HexToAddress(account.Address),
		PublicKey:           hexutil.MustDecode(account.PublicKey),
		CompressedPublicKey: hexutil.MustDecode(account.CompressedPublicKey),
		TenantID:            account.TenantID,
		OwnerID:             account.OwnerID,
		StoreID:             account.StoreID,
		Attributes:          account.Attributes,
		CreatedAt:           account.CreatedAt,
		UpdatedAt:           account.UpdatedAt,
	}
}

func NewAccountEntityArr(accounts []*models.Account) []*entities.Account {
	res := []*entities.Account{}
	for _, acc := range accounts {
		mAcc := NewAccountEntity(acc)
		res = append(res, mAcc)
	}
	return res
}
