package models

import (
	"time"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type Faucet struct {
	tableName struct{} `pg:"faucets"` // nolint:unused,structcheck // reason

	UUID            string `pg:",pk"`
	Name            string
	TenantID        string
	ChainRule       string
	CreditorAccount string
	MaxBalance      string
	Amount          string
	Cooldown        string
	CreatedAt       time.Time `pg:"default:now()"`
	UpdatedAt       time.Time `pg:"default:now()"`
}

func NewFaucet(faucet *entities.Faucet) *Faucet {
	return &Faucet{
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
}

func NewFaucets(faucets []*Faucet) []*entities.Faucet {
	res := []*entities.Faucet{}
	for _, f := range faucets {
		res = append(res, f.ToEntity())
	}

	return res
}

func (faucet *Faucet) ToEntity() *entities.Faucet {
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
