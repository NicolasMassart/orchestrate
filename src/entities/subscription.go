package entities

import (
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

type Subscription struct {
	UUID            string
	Address         ethcommon.Address
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
