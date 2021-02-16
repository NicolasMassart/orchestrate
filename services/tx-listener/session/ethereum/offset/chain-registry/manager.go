package chainregistry

import (
	"context"
	"fmt"
	"sync"

	"github.com/containous/traefik/v2/pkg/log"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/errors"
	registry "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/client"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/tx-listener/dynamic"
)

const component = "tx-listener.offset"

type Manager struct {
	sm       *sync.Map
	registry registry.ChainRegistryClient
}

func NewManager(r registry.ChainRegistryClient) *Manager {
	return &Manager{
		sm:       &sync.Map{},
		registry: r,
	}
}

func (m *Manager) GetLastBlockNumber(ctx context.Context, chain *dynamic.Chain) (uint64, error) {
	n, err := m.registry.GetChainByUUID(ctx, chain.UUID)
	if err != nil {
		return 0, errors.FromError(err).ExtendComponent(component)
	}
	return *n.ListenerCurrentBlock, nil
}

func (m *Manager) SetLastBlockNumber(ctx context.Context, chain *dynamic.Chain, blockNumber uint64) error {
	if chain.Listener.CurrentBlock == blockNumber {
		log.FromContext(ctx).WithField("block_number", blockNumber).
			Warn("ignored set last block number. Chain is already at same block")
		return nil
	}
	err := m.registry.UpdateBlockPosition(ctx, chain.UUID, blockNumber)
	if err != nil {
		return errors.FromError(err).ExtendComponent(component)
	}
	return nil
}

func (m *Manager) GetLastTxIndex(_ context.Context, chain *dynamic.Chain, blockNumber uint64) (uint64, error) {
	txIndex, ok := m.sm.Load(fmt.Sprintf("txIndex-%v-%v", chain.UUID, blockNumber))
	if !ok {
		return 0, nil
	}
	return txIndex.(uint64), nil
}

func (m *Manager) SetLastTxIndex(_ context.Context, chain *dynamic.Chain, blockNumber, txIndex uint64) error {
	m.sm.Store(fmt.Sprintf("txIndex-%v-%v", chain.UUID, blockNumber), txIndex)
	return nil
}
