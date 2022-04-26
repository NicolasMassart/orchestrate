package inmemory

import (
	"context"
	"sync"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	"github.com/ethereum/go-ethereum/common"
)

type retrySessionInMemory struct {
	indexBySessID         map[string]*entities.Job   // sessID => Job
	indexByTxHash         map[string]string          // TxHash => SessID
	aggregatedByChainUUID map[string]map[string]bool // ChainUUID => SessID => Bool
	mux                   *sync.RWMutex
}

func NewRetrySessionInMemory() store.RetrySessions {
	return &retrySessionInMemory{
		indexBySessID:         make(map[string]*entities.Job),
		indexByTxHash:         make(map[string]string),
		aggregatedByChainUUID: make(map[string]map[string]bool),
		mux:                   &sync.RWMutex{},
	}
}

func (m *retrySessionInMemory) Add(_ context.Context, sessID string, job *entities.Job) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexBySessID[sessID]; ok {
		return errors.AlreadyExistsError("retry session %q is duplicated", sessID)
	}

	m.indexBySessID[sessID] = job
	m.indexByTxHash[job.Transaction.Hash.String()] = sessID
	if _, ok := m.aggregatedByChainUUID[job.ChainUUID]; !ok {
		m.aggregatedByChainUUID[job.ChainUUID] = make(map[string]bool)
	}
	m.aggregatedByChainUUID[job.ChainUUID][sessID] = true
	return nil
}

func (m *retrySessionInMemory) Remove(_ context.Context, sessID string) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexBySessID[sessID]; !ok {
		return errors.NotFoundError("retry session %q is not found", sessID)
	}

	job := m.indexBySessID[sessID]
	delete(m.aggregatedByChainUUID[job.ChainUUID], sessID)
	delete(m.indexByTxHash, job.Transaction.Hash.String())
	delete(m.indexBySessID, sessID)
	return nil
}

func (m *retrySessionInMemory) Has(_ context.Context, sessID string) bool {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexBySessID[sessID]; !ok {
		return false
	}

	return true
}

func (m *retrySessionInMemory) GetByTxHash(_ context.Context, chainUUID string, txHash *common.Hash) (string, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	if sessID, ok := m.indexByTxHash[txHash.String()]; ok {
		job := m.indexBySessID[sessID]
		if chainUUID == "" || job.ChainUUID == chainUUID {
			return sessID, nil
		}
	}

	return "", errors.NotFoundError("retry session with hash %q is not found", txHash.String())
}

func (m *retrySessionInMemory) ListByChainUUID(_ context.Context, chainUUID string) ([]string, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	sessionIDs := []string{}
	if _, ok := m.aggregatedByChainUUID[chainUUID]; !ok {
		return sessionIDs, errors.NotFoundError("there is not retry sessions in chain %q", chainUUID)
	}

	for jobUUID := range m.aggregatedByChainUUID[chainUUID] {
		sessID := m.indexByTxHash[jobUUID]
		// @TODO: Investigate why delete() leaves empty values
		if sessID != "" {
			sessionIDs = append(sessionIDs, sessID)
		}
	}

	return sessionIDs, nil
}

func (m *retrySessionInMemory) DeletePerChainUUID(ctx context.Context, chainUUID string) error {
	if _, ok := m.aggregatedByChainUUID[chainUUID]; !ok {
		return errors.NotFoundError("there is not retry sessions in chain %q", chainUUID)
	}

	for sessID := range m.aggregatedByChainUUID[chainUUID] {
		_ = m.Remove(ctx, sessID)
	}
	return nil
}
