package types

import (
	"time"

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

type AccountResponse struct {
	Alias               string            `json:"alias" example:"personal-account"`                                                                                                                                              // Alias of the account.
	Address             string            `json:"address" example:"0x1abae27a0cbfb02945720425d3b80c7e09728534" swaggertype:"string"`                                                                                             // Address of the account.
	PublicKey           string            `json:"publicKey" example:"0x048e66b3e549818ea2cb354fb70749f6c8de8fa484f7530fc447d5fe80a1c424e4f5ae648d648c980ae7095d1efad87161d83886ca4b6c498ac22a93da5099014a" swaggertype:"string"` // Public key of the account.
	CompressedPublicKey string            `json:"compressedPublicKey" example:"0x048e66b3e549818ea2cb354fb70749f6c8de8fa484f7530fc447" swaggertype:"string"`                                                                     // Compressed public key of the account.
	TenantID            string            `json:"tenantID" example:"tenantFoo"`                                                                                                                                                  // ID of the tenant executing the API.
	OwnerID             string            `json:"ownerID,omitempty" example:"foo"`                                                                                                                                               // ID of the account owner.
	StoreID             string            `json:"storeID,omitempty" example:"myQKMStoreID"`                                                                                                                                      // ID of the Quorum Key Manager store containing the account.
	Attributes          map[string]string `json:"attributes,omitempty"`                                                                                                                                                          // Additional information attached to the account.
	CreatedAt           time.Time         `json:"createdAt" example:"2020-07-09T12:35:42.115395Z"`                                                                                                                               // Date and time at which the account was created.
	UpdatedAt           time.Time         `json:"updatedAt,omitempty" example:"2020-07-09T12:35:42.115395Z"`                                                                                                                     // Date and time at which the account details were updated.
}
