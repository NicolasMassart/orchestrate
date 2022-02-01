package types

import (
	"github.com/consensys/quorum-key-manager/src/stores/api/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type CreateAccountRequest struct {
	Alias      string            `json:"alias" validate:"omitempty" example:"personal-account" ` // Alias of the account.
	Chain      string            `json:"chain" validate:"omitempty" example:"besu"`              // Name of the chain. This value should match the chain name defined in the chain creation.
	StoreID    string            `json:"storeID" validate:"omitempty" example:"qkmStoreID"`      // ID of the Quorum Key Manager store containing the account.
	Attributes map[string]string `json:"attributes,omitempty"`                                   // Additional information attached to the account.
}

type ImportAccountRequest struct {
	Alias      string            `json:"alias" validate:"omitempty" example:"personal-account"`                                                                            // Alias of the account.
	Chain      string            `json:"chain" validate:"omitempty" example:"quorum"`                                                                                      // Name of the chain. This value should match the chain name defined in the chain creation.
	PrivateKey hexutil.Bytes     `json:"privateKey" validate:"required" example:"0x66232652FDFFD802B7252A456DBD8F3ECC0352BBDE76C23B40AFE8AEBD714E2D" swaggertype:"string"` // Private key of the account.
	StoreID    string            `json:"storeID" validate:"omitempty" example:"qkmStoreID"`                                                                                // ID of the Quorum Key Manager store containing the account.
	Attributes map[string]string `json:"attributes,omitempty"`                                                                                                             // Additional information attached to the account.
}

type UpdateAccountRequest struct {
	Alias      string            `json:"alias" validate:"omitempty"  example:"personal-account"` // Alias of the account.
	StoreID    string            `json:"storeID" validate:"omitempty" example:"qkmStoreID"`      // ID of the Quorum Key Manager store containing the account.
	Attributes map[string]string `json:"attributes,omitempty"`                                   // Additional information attached to the account.
}

type SignMessageRequest struct {
	types.SignMessageRequest
	StoreID string `json:"storeID" validate:"omitempty" example:"qkmStoreID"` // ID of the Quorum Key Manager store containing the account.
}

type SignTypedDataRequest struct {
	types.SignTypedDataRequest
	StoreID string `json:"storeID" validate:"omitempty" example:"qkmStoreID"` // ID of the Quorum Key Manager store containing the account.
}
