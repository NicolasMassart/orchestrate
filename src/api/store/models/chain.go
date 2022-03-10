package models

import (
	"math/big"
	"time"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
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

func NewChain(chain *entities.Chain) *Chain {
	chainModel := &Chain{
		UUID:                      chain.UUID,
		Name:                      chain.Name,
		TenantID:                  chain.TenantID,
		OwnerID:                   chain.OwnerID,
		URLs:                      chain.URLs,
		ListenerDepth:             chain.ListenerDepth,
		ListenerCurrentBlock:      chain.ListenerCurrentBlock,
		ListenerStartingBlock:     chain.ListenerStartingBlock,
		ListenerExternalTxEnabled: chain.ListenerExternalTxEnabled,
		PrivateTxManagerURL:       chain.PrivateTxManagerURL,
		Labels:                    chain.Labels,
		CreatedAt:                 chain.CreatedAt,
		UpdatedAt:                 chain.UpdatedAt,
	}

	if chain.ListenerBackOffDuration.Milliseconds() > 0 {
		chainModel.ListenerBackOffDuration = chain.ListenerBackOffDuration.String()
	}

	if chain.ChainID != nil {
		chainModel.ChainID = chain.ChainID.String()
	}

	return chainModel
}

func NewChains(chains []*Chain) []*entities.Chain {
	res := []*entities.Chain{}
	for _, c := range chains {
		res = append(res, c.ToEntity())
	}

	return res
}

func (c *Chain) ToEntity() *entities.Chain {
	listenerBackOffDuration, _ := time.ParseDuration(c.ListenerBackOffDuration)

	chain := &entities.Chain{
		UUID:                      c.UUID,
		Name:                      c.Name,
		TenantID:                  c.TenantID,
		OwnerID:                   c.OwnerID,
		URLs:                      c.URLs,
		ChainID:                   (*big.Int)(utils.StringBigIntToHex(c.ChainID)),
		ListenerDepth:             c.ListenerDepth,
		ListenerCurrentBlock:      c.ListenerCurrentBlock,
		ListenerStartingBlock:     c.ListenerStartingBlock,
		ListenerBackOffDuration:   listenerBackOffDuration,
		ListenerExternalTxEnabled: c.ListenerExternalTxEnabled,
		PrivateTxManagerURL:       c.PrivateTxManagerURL,
		Labels:                    c.Labels,
		CreatedAt:                 c.CreatedAt,
		UpdatedAt:                 c.UpdatedAt,
	}

	return chain
}
