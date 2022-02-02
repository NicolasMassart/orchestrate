package parsers

import (
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

func NewFaucetEntity(faucet *models.Faucet) *entities.Faucet {
	return &entities.Faucet{
		UUID:            faucet.UUID,
		Name:            faucet.Name,
		TenantID:        faucet.TenantID,
		ChainRule:       faucet.ChainRule,
		CreditorAccount: ethcommon.HexToAddress(faucet.CreditorAccount),
		MaxBalance:      *utils.StringBigIntToHex(faucet.MaxBalance),
		Amount:          *utils.StringBigIntToHex(faucet.Amount),
		Cooldown:        faucet.Cooldown,
		CreatedAt:       faucet.CreatedAt,
		UpdatedAt:       faucet.UpdatedAt,
	}
}

func NewFaucetEntityArr(faucets []*models.Faucet) []*entities.Faucet {
	res := []*entities.Faucet{}
	for _, f := range faucets {
		res = append(res, NewFaucetEntity(f))
	}

	return res
}

func NewFaucetModel(faucet *entities.Faucet) *models.Faucet {
	f := &models.Faucet{
		UUID:            faucet.UUID,
		Name:            faucet.Name,
		TenantID:        faucet.TenantID,
		ChainRule:       faucet.ChainRule,
		CreditorAccount: faucet.CreditorAccount.Hex(),
		MaxBalance:      faucet.MaxBalance.ToInt().String(),
		Amount:          faucet.Amount.ToInt().String(),
		Cooldown:        faucet.Cooldown,
		CreatedAt:       faucet.CreatedAt,
	}

	return f
}
