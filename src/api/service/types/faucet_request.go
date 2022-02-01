package types

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type RegisterFaucetRequest struct {
	Name            string            `json:"name" validate:"required" example:"faucet-mainnet"`                                                             // Name of the faucet.
	ChainRule       string            `json:"chainRule" validate:"required" example:"mainnet"`                                                               // Name of the chain on which to register the faucet.
	CreditorAccount ethcommon.Address `json:"creditorAccount" validate:"required" example:"0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18" swaggertype:"string"` // Address of the faucet creditor's account.
	MaxBalance      hexutil.Big       `json:"maxBalance" validate:"required" example:"0x254582f40" swaggertype:"string"`                                     // Expected maximum balance of beneficiary. It won't fund past this value.
	Amount          hexutil.Big       `json:"amount" validate:"required" example:"0xF4240" swaggertype:"string"`                                             // Amount, in Wei, sent to the beneficiary on each funding transaction.
	Cooldown        string            `json:"cooldown" validate:"required,isDuration" example:"10s"`                                                         // Waiting time in between two funding transactions to the same beneficiary.
}

type UpdateFaucetRequest struct {
	Name            string            `json:"name,omitempty" validate:"omitempty" example:"faucet-mainnet"`                                                             // Name of the faucet.
	ChainRule       string            `json:"chainRule,omitempty" validate:"omitempty" example:"mainnet"`                                                               // Name of the chain on which to register the faucet.
	CreditorAccount ethcommon.Address `json:"creditorAccount,omitempty" validate:"omitempty" example:"0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18" swaggertype:"string"` // Address of the faucet creditor's account.
	MaxBalance      hexutil.Big       `json:"maxBalance,omitempty" validate:"omitempty" example:"0x254582f40" swaggertype:"string"`                                     // Expected maximum balance of beneficiary. It won't fund past this value.
	Amount          hexutil.Big       `json:"amount,omitempty" validate:"omitempty" example:"0x254582f40" swaggertype:"string"`                                         // Amount, in Wei, sent to the beneficiary on each funding transaction.
	Cooldown        string            `json:"cooldown,omitempty" validate:"omitempty,isDuration" example:"10s"`                                                         // Waiting time in between two funding transactions to the same beneficiary.
}
