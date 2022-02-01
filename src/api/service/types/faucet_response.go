package types

import (
	"encoding/json"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type FaucetResponse struct {
	UUID            string            `json:"uuid" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`                                        // UUID of the faucet.
	Name            string            `json:"name" validate:"required" example:"faucet-mainnet"`                                          // Name of the faucet.
	TenantID        string            `json:"tenantID,omitempty" example:"foo"`                                                           // ID of the tenant executing the API.
	ChainRule       string            `json:"chainRule,omitempty" example:"mainnet"`                                                      // Name of the chain on which the faucet is registered.
	CreditorAccount ethcommon.Address `json:"creditorAccount"  example:"0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18" swaggertype:"string"` // Address of the faucet creditor's account.
	MaxBalance      hexutil.Big       `json:"maxBalance,omitempty" validate:"required" example:"0x16345785D8A0000" swaggertype:"string"`  // Expected maximum balance of beneficiary. It won't fund past this value.
	Amount          hexutil.Big       `json:"amount,omitempty" validate:"required" example:"0xD529AE9E860000" swaggertype:"string"`       // Amount, in Wei, sent to the beneficiary on each funding transaction.
	Cooldown        string            `json:"cooldown,omitempty" validate:"required,isDuration" example:"10s"`                            // Waiting time in between two funding transactions to the same beneficiary.
	CreatedAt       time.Time         `json:"createdAt" example:"2020-07-09T12:35:42.115395Z"`                                            // Date and time at which the faucet was registered.
	UpdatedAt       time.Time         `json:"updatedAt" example:"2020-07-09T12:35:42.115395Z"`                                            // Date and time at which the faucet details were updated.
}

type faucetResponseJSON struct {
	UUID            string    `json:"uuid"`
	Name            string    `json:"name"`
	TenantID        string    `json:"tenantID"`
	ChainRule       string    `json:"chainRule,omitempty"`
	CreditorAccount string    `json:"creditorAccount"`
	MaxBalance      string    `json:"maxBalance,omitempty"`
	Amount          string    `json:"amount,omitempty"`
	Cooldown        string    `json:"cooldown,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt,omitempty"`
}

func (a *FaucetResponse) MarshalJSON() ([]byte, error) {
	res := &faucetResponseJSON{
		UUID:            a.UUID,
		Name:            a.Name,
		TenantID:        a.TenantID,
		ChainRule:       a.ChainRule,
		CreditorAccount: a.CreditorAccount.String(),
		MaxBalance:      a.MaxBalance.String(),
		Amount:          a.Amount.String(),
		Cooldown:        a.Cooldown,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}

	return json.Marshal(res)
}
