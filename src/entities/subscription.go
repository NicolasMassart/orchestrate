package entities

import (
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// @TODO Support subscription topics (ABI methods)
type Subscription struct {
	UUID            string
	ContractAddress ethcommon.Address
	ChainUUID       string
	ContractName    string
	ContractTag     string
	EventStreamUUID string
	FromBlock       *uint64
	TenantID        string
	OwnerID         string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
