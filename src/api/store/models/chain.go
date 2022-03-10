package models

import (
	"time"
)

type Chain struct {
	tableName struct{} `pg:"chains"` // nolint:unused,structcheck // reason

	UUID                      string `pg:",pk"`
	Name                      string
	TenantID                  string
	OwnerID                   string
	URLs                      []string `pg:"urls,array"`
	PrivateTxManagerURL       string   `pg:"private_tx_manager_url"`
	ChainID                   string
	ListenerDepth             uint64
	ListenerCurrentBlock      uint64
	ListenerStartingBlock     uint64
	ListenerBackOffDuration   string
	ListenerExternalTxEnabled bool
	Labels                    map[string]string
	CreatedAt                 time.Time `pg:"default:now()"`
	UpdatedAt                 time.Time `pg:"default:now()"`
}
