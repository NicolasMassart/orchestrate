package inmemory

import (
	"context"
	"sync"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/store"
)

type chainState struct {
	activeChains map[string]*entities.Chain
	mux          *sync.RWMutex
}

func NewChainState() store.Chain {
	return &chainState{
		activeChains: make(map[string]*entities.Chain),
		mux:          &sync.RWMutex{},
	}
}

func (m *chainState) Add(_ context.Context, chain *entities.Chain) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.activeChains[chain.UUID]; ok {
		return errors.AlreadyExistsError("chain %q is duplicated", chain.UUID)
	}

	m.activeChains[chain.UUID] = chain
	return nil
}

func (m *chainState) Get(_ context.Context, chainUUID string) (*entities.Chain, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	if chain, ok := m.activeChains[chainUUID]; ok && chain.ChainID != nil {
		return chain, nil
	}

	return nil, errors.NotFoundError("chain %q is not found", chainUUID)
}

func (m *chainState) Update(_ context.Context, chain *entities.Chain) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.activeChains[chain.UUID]; !ok {
		return errors.NotFoundError("chain %q is not found", chain.UUID)
	}

	m.activeChains[chain.UUID] = chain
	return nil
}

func (m *chainState) Delete(_ context.Context, chainUUID string) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.activeChains[chainUUID]; !ok {
		return errors.NotFoundError("chain %q is not found", chainUUID)
	}

	delete(m.activeChains, chainUUID)
	return nil
}
