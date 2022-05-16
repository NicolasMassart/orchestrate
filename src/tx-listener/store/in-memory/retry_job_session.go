package inmemory

import (
	"context"
	"sync"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	"github.com/ethereum/go-ethereum/common"
)

type retryJobSessionState struct {
	indexByJobUUID        map[string]*entities.Job   // jobUUID => Job
	indexByTxHash         map[string]string          // TxHash => jobUUID
	aggregatedByChainUUID map[string]map[string]bool // ChainUUID => jobUUID => Bool
	mux                   *sync.RWMutex
}

func NewRetryJobSessionState() store.RetryJobSession {
	return &retryJobSessionState{
		indexByJobUUID:        make(map[string]*entities.Job),
		indexByTxHash:         make(map[string]string),
		aggregatedByChainUUID: make(map[string]map[string]bool),
		mux:                   &sync.RWMutex{},
	}
}

func (m *retryJobSessionState) Add(_ context.Context, job *entities.Job) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexByJobUUID[job.UUID]; ok {
		return errors.AlreadyExistsError("retry job session %q is duplicated", job.UUID)
	}

	m.indexByJobUUID[job.UUID] = job
	m.indexByTxHash[job.Transaction.Hash.String()] = job.UUID
	if _, ok := m.aggregatedByChainUUID[job.ChainUUID]; !ok {
		m.aggregatedByChainUUID[job.ChainUUID] = make(map[string]bool)
	}
	m.aggregatedByChainUUID[job.ChainUUID][job.UUID] = true
	return nil
}

func (m *retryJobSessionState) Remove(_ context.Context, jobUUID string) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexByJobUUID[jobUUID]; !ok {
		return errors.NotFoundError("retry session %q is not found", jobUUID)
	}

	job := m.indexByJobUUID[jobUUID]
	delete(m.aggregatedByChainUUID[job.ChainUUID], jobUUID)
	delete(m.indexByTxHash, job.Transaction.Hash.String())
	delete(m.indexByJobUUID, jobUUID)
	return nil
}

func (m *retryJobSessionState) Has(_ context.Context, jobUUID string) bool {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexByJobUUID[jobUUID]; !ok {
		return false
	}

	return true
}

func (m *retryJobSessionState) GetByTxHash(_ context.Context, chainUUID string, txHash *common.Hash) (string, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	if jobUUID, ok := m.indexByTxHash[txHash.String()]; ok {
		job := m.indexByJobUUID[jobUUID]
		if chainUUID == "" || job.ChainUUID == chainUUID {
			return jobUUID, nil
		}
	}

	return "", errors.NotFoundError("retry session with hash %q is not found", txHash.String())
}

func (m *retryJobSessionState) ListByChainUUID(_ context.Context, chainUUID string) ([]string, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	jobUUIDs := []string{}
	if _, ok := m.aggregatedByChainUUID[chainUUID]; !ok {
		return jobUUIDs, errors.NotFoundError("there is not retry sessions in chain %q", chainUUID)
	}

	for jobUUID := range m.aggregatedByChainUUID[chainUUID] {
		if jobUUID != "" {
			jobUUIDs = append(jobUUIDs, jobUUID)
		}
	}

	return jobUUIDs, nil
}

func (m *retryJobSessionState) DeletePerChainUUID(ctx context.Context, chainUUID string) error {
	if _, ok := m.aggregatedByChainUUID[chainUUID]; !ok {
		return errors.NotFoundError("there is not retry sessions in chain %q", chainUUID)
	}

	for jobUUID := range m.aggregatedByChainUUID[chainUUID] {
		_ = m.Remove(ctx, jobUUID)
	}
	return nil
}
