package inmemory

import (
	"context"
	"sync"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	"github.com/ethereum/go-ethereum/common"
)

type subscriptionState struct {
	indexedByUUID          map[string]*entities.Subscription
	indexedByAddress       map[string]map[string]bool
	indexedAddrByChainUUID map[string]map[string]int
	mux                    *sync.RWMutex
}

func NewSubscriptionState() store.Subscriptions {
	return &subscriptionState{
		indexedByUUID:          make(map[string]*entities.Subscription), // SubUUID => Subscription
		indexedByAddress:       make(map[string]map[string]bool),        // Address => []SubUUID => bool
		indexedAddrByChainUUID: make(map[string]map[string]int),         // ChainUUID => []Address => counter
		mux:                    &sync.RWMutex{},
	}
}

func (m *subscriptionState) Add(_ context.Context, sub *entities.Subscription) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, ok := m.indexedByUUID[sub.UUID]; ok {
		return errors.AlreadyExistsError("subscription %q is duplicated", sub.UUID)
	}

	if _, ok := m.indexedByAddress[sub.ContractAddress.String()]; !ok {
		m.indexedByAddress[sub.ContractAddress.String()] = make(map[string]bool)
	}

	m.indexedByAddress[sub.ContractAddress.String()][sub.UUID] = true

	if _, ok := m.indexedAddrByChainUUID[sub.ChainUUID]; !ok {
		m.indexedAddrByChainUUID[sub.ChainUUID] = make(map[string]int)
	}

	if _, ok := m.indexedAddrByChainUUID[sub.ChainUUID][sub.ContractAddress.String()]; !ok {
		m.indexedAddrByChainUUID[sub.ChainUUID][sub.ContractAddress.String()] = 0
	}

	m.indexedAddrByChainUUID[sub.ChainUUID][sub.ContractAddress.String()]++
	return nil
}

func (m *subscriptionState) Update(ctx context.Context, sub *entities.Subscription) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexedByUUID[sub.UUID]; !ok {
		return errors.NotFoundError("subscription %q is not found", sub.UUID)
	}

	m.indexedByUUID[sub.UUID] = sub
	return nil
}

func (m *subscriptionState) Remove(_ context.Context, subUUID string) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexedByUUID[subUUID]; !ok {
		return errors.NotFoundError("subscription %q is not found", subUUID)
	}

	sub := m.indexedByUUID[subUUID]
	m.indexedAddrByChainUUID[sub.ChainUUID][sub.ContractAddress.String()]--
	delete(m.indexedAddrByChainUUID[sub.ChainUUID], sub.UUID)
	delete(m.indexedByAddress, sub.UUID)
	delete(m.indexedByUUID, sub.UUID)
	return nil
}

func (m *subscriptionState) ListAddressesPerChainUUID(ctx context.Context, chainUUID string) ([]common.Address, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	addrs := []common.Address{}
	if _, ok := m.indexedAddrByChainUUID[chainUUID]; !ok {
		return addrs, nil
	}

	for addr, counter := range m.indexedAddrByChainUUID[chainUUID] {
		if counter > 0 {
			addrs = append(addrs, common.HexToAddress(addr))
		}
	}

	return addrs, nil
}

func (m *subscriptionState) ListPerChainUUID(_ context.Context, chainUUID string) ([]*entities.Subscription, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	subscriptions := []*entities.Subscription{}
	if _, ok := m.indexedAddrByChainUUID[chainUUID]; !ok {
		return nil, nil
	}

	if chainAddrs, ok := m.indexedAddrByChainUUID[chainUUID]; ok {
		for addr := range chainAddrs {
			if subUUIDs, ok := m.indexedByAddress[addr]; ok {
				for subUUID := range subUUIDs {
					subscriptions = append(subscriptions, m.indexedByUUID[subUUID])
				}
			}
		}
	}

	return subscriptions, nil
}
