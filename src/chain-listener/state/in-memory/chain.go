package inmemory

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/chain-listener/state"
	"github.com/consensys/orchestrate/src/entities"
)

type chainInMemory struct {
	activeChains map[string]*entities.Chain
}

func NewChainInMemory() state.Chain {
	return &chainInMemory{
		activeChains: make(map[string]*entities.Chain),
	}
}

func (m *chainInMemory) Add(_ context.Context, chain *entities.Chain) error {
	// @TODO Mutex
	if _, ok := m.activeChains[chain.UUID]; ok {
		return errors.AlreadyExistsError("chain %q is duplicated", chain.UUID)
	}

	m.activeChains[chain.UUID] = chain
	return nil
}

func (m *chainInMemory) Get(_ context.Context, chainUUID string) (*entities.Chain, error) {
	// @TODO Mutex
	if chain, ok := m.activeChains[chainUUID]; ok {
		return chain, nil
	}

	return nil, errors.NotFoundError("chain %q is not found", chainUUID)
}

func (m *chainInMemory) Update(_ context.Context, chain *entities.Chain) error {
	// @TODO Mutex
	if _, ok := m.activeChains[chain.UUID]; !ok {
		return errors.NotFoundError("chain %q is not found", chain.UUID)
	}

	m.activeChains[chain.UUID] = chain
	return nil
}

func (m *chainInMemory) Remove(_ context.Context, chainUUID string) error {
	// @TODO Mutex
	if _, ok := m.activeChains[chainUUID]; !ok {
		return errors.NotFoundError("chain %q is not found", chainUUID)
	}

	delete(m.activeChains, chainUUID)
	return nil
}
