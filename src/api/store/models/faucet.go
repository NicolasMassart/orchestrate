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

func NewFaucet(f *entities.Faucet) *Faucet {
	return &Faucet{
		UUID:            f.UUID,
		Name:            f.Name,
		TenantID:        f.TenantID,
		ChainRule:       f.ChainRule,
		CreditorAccount: f.CreditorAccount.Hex(),
		MaxBalance:      f.MaxBalance.ToInt().String(),
		Amount:          f.Amount.ToInt().String(),
		Cooldown:        f.Cooldown,
		CreatedAt:       f.CreatedAt,
		UpdatedAt:       f.UpdatedAt,
	}
}

func NewFaucets(faucets []*Faucet) []*entities.Faucet {
	res := []*entities.Faucet{}
	for _, f := range faucets {
		res = append(res, f.ToEntity())
	}

	return res
}

func (f *Faucet) ToEntity() *entities.Faucet {
	return &entities.Faucet{
		UUID:            f.UUID,
		Name:            f.Name,
		TenantID:        f.TenantID,
		ChainRule:       f.ChainRule,
		CreditorAccount: ethcommon.HexToAddress(f.CreditorAccount),
		MaxBalance:      *utils.StringBigIntToHex(f.MaxBalance),
		Amount:          *utils.StringBigIntToHex(f.Amount),
		Cooldown:        f.Cooldown,
		CreatedAt:       f.CreatedAt,
		UpdatedAt:       f.UpdatedAt,
	}
}
