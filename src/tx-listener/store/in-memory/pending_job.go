package inmemory

import (
	"context"
	"sync"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	"github.com/ethereum/go-ethereum/common"
)

type pendingJobInMemory struct {
	indexedByJobUUID   map[string]*entities.Job
	indexedByTxHash    map[string]*entities.Job
	indexedByChainUUID map[string]map[string]bool
	mux                *sync.RWMutex
}

func NewPendingJobInMemory() store.PendingJob {
	return &pendingJobInMemory{
		indexedByJobUUID:   make(map[string]*entities.Job),   // JobUUID => Job
		indexedByTxHash:    make(map[string]*entities.Job),   // TxHash => Job
		indexedByChainUUID: make(map[string]map[string]bool), // ChainUUID => JobUUID => bool
		mux:                &sync.RWMutex{},
	}
}

func (m *pendingJobInMemory) Add(_ context.Context, job *entities.Job) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexedByJobUUID[job.UUID]; ok {
		return errors.AlreadyExistsError("job %q is duplicated", job.UUID)
	}

	m.indexedByJobUUID[job.UUID] = job
	m.indexedByTxHash[job.Transaction.Hash.String()] = job
	if _, ok := m.indexedByChainUUID[job.ChainUUID]; !ok {
		m.indexedByChainUUID[job.ChainUUID] = make(map[string]bool)
	}
	m.indexedByChainUUID[job.ChainUUID][job.UUID] = true
	return nil
}

func (m *pendingJobInMemory) Update(_ context.Context, job *entities.Job) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexedByJobUUID[job.UUID]; !ok {
		return errors.NotFoundError("job %q is duplicated", job.UUID)
	}

	previousJob := m.indexedByJobUUID[job.UUID]
	m.indexedByJobUUID[job.UUID] = job
	delete(m.indexedByTxHash, previousJob.Transaction.Hash.String())
	m.indexedByTxHash[job.Transaction.Hash.String()] = job
	return nil
}

func (m *pendingJobInMemory) Remove(_ context.Context, jobUUID string) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexedByJobUUID[jobUUID]; !ok {
		return errors.NotFoundError("job %q is not found", jobUUID)
	}

	job := m.indexedByJobUUID[jobUUID]
	delete(m.indexedByChainUUID[job.ChainUUID], job.UUID)
	delete(m.indexedByTxHash, job.Transaction.Hash.String())
	delete(m.indexedByJobUUID, job.UUID)
	return nil
}

func (m *pendingJobInMemory) GetByTxHash(_ context.Context, chainUUID string, txHash *common.Hash) (*entities.Job, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	if job, ok := m.indexedByTxHash[txHash.String()]; ok {
		if chainUUID == "" || job.ChainUUID == chainUUID {
			return job, nil
		}
	}

	return nil, errors.NotFoundError("job with hash %q is not found", txHash.String())
}

func (m *pendingJobInMemory) GetJobUUID(_ context.Context, jobUUID string) (*entities.Job, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	if _, ok := m.indexedByJobUUID[jobUUID]; !ok {
		return nil, errors.NotFoundError("job %q is not found", jobUUID)
	}
	return m.indexedByJobUUID[jobUUID], nil
}

func (m *pendingJobInMemory) ListPerChainUUID(_ context.Context, chainUUID string) ([]*entities.Job, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	jobs := []*entities.Job{}
	if _, ok := m.indexedByChainUUID[chainUUID]; !ok {
		return nil, nil
	}

	for jobUUID := range m.indexedByChainUUID[chainUUID] {
		jobs = append(jobs, m.indexedByJobUUID[jobUUID])
	}

	return jobs, nil
}

func (m *pendingJobInMemory) ListChainUUID(_ context.Context, chainUUID string) ([]string, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	chainUUIDs := []string{}
	for uuid := range m.indexedByChainUUID {
		chainUUIDs = append(chainUUIDs, uuid)
	}

	return chainUUIDs, nil
}

func (m *pendingJobInMemory) DeletePerChainUUID(ctx context.Context, chainUUID string) error {
	// @TODO Mutex
	if _, ok := m.indexedByChainUUID[chainUUID]; !ok {
		return errors.NotFoundError("there is not jobs in chain %q", chainUUID)
	}

	for jobUUID := range m.indexedByChainUUID[chainUUID] {
		_ = m.Remove(ctx, jobUUID)
	}
	return nil
}
