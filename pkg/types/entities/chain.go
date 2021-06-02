package entities

import "time"

type Chain struct {
	UUID                      string
	Name                      string
	TenantID                  string
	URLs                      []string
	ChainID                   string
	ListenerDepth             uint64
	ListenerCurrentBlock      uint64
	ListenerStartingBlock     uint64
	ListenerBackOffDuration   string
	ListenerExternalTxEnabled bool
	PrivateTxManager          *PrivateTxManager
	Labels                    map[string]string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}
