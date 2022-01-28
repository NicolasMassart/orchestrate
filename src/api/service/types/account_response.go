package types

import (
	"encoding/json"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type AccountResponse struct {
	Alias               string            `json:"alias" example:"personal-account"`                                                                                                                                              // Alias of the account.
	Address             ethcommon.Address `json:"address" example:"0x1abae27a0cbfb02945720425d3b80c7e09728534" swaggertype:"string"`                                                                                             // Address of the account.
	PublicKey           hexutil.Bytes     `json:"publicKey" example:"0x048e66b3e549818ea2cb354fb70749f6c8de8fa484f7530fc447d5fe80a1c424e4f5ae648d648c980ae7095d1efad87161d83886ca4b6c498ac22a93da5099014a" swaggertype:"string"` // Public key of the account.
	CompressedPublicKey hexutil.Bytes     `json:"compressedPublicKey" example:"0x048e66b3e549818ea2cb354fb70749f6c8de8fa484f7530fc447" swaggertype:"string"`                                                                     // Compressed public key of the account.
	TenantID            string            `json:"tenantID" example:"tenantFoo"`                                                                                                                                                  // ID of the tenant executing the API.
	OwnerID             string            `json:"ownerID,omitempty" example:"foo"`                                                                                                                                               // ID of the account owner.
	StoreID             string            `json:"storeID,omitempty" example:"myQKMStoreID"`                                                                                                                                      // ID of the Quorum Key Manager store containing the account.
	Attributes          map[string]string `json:"attributes,omitempty"`                                                                                                                                                          // Additional information attached to the account.
	CreatedAt           time.Time         `json:"createdAt" example:"2020-07-09T12:35:42.115395Z"`                                                                                                                               // Date and time at which the account was created.
	UpdatedAt           time.Time         `json:"updatedAt,omitempty" example:"2020-07-09T12:35:42.115395Z"`                                                                                                                     // Date and time at which the account details were updated.
}

type accountResponseJSON struct {
	Alias               string            `json:"alias"`
	Address             string            `json:"address"`
	PublicKey           string            `json:"publicKey"`
	CompressedPublicKey string            `json:"compressedPublicKey"`
	TenantID            string            `json:"tenantID"`
	OwnerID             string            `json:"ownerID,omitempty"`
	StoreID             string            `json:"storeID,omitempty"`
	Attributes          map[string]string `json:"attributes,omitempty"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt,omitempty"`
}

func (a *AccountResponse) MarshalJSON() ([]byte, error) {
	res := &accountResponseJSON{
		Alias:               a.Alias,
		PublicKey:           a.PublicKey.String(),
		CompressedPublicKey: a.CompressedPublicKey.String(),
		Address:             a.Address.Hex(),
		TenantID:            a.TenantID,
		OwnerID:             a.OwnerID,
		StoreID:             a.StoreID,
		Attributes:          a.Attributes,
		CreatedAt:           a.CreatedAt,
		UpdatedAt:           a.UpdatedAt,
	}

	return json.Marshal(res)
}
