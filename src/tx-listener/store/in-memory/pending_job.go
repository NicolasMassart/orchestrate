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
	indexByJobUUID        map[string]*entities.Job
	indexByTxHash         map[string]*entities.Job
	aggregatedByChainUUID map[string]map[string]bool
	mux                   *sync.RWMutex
}

func NewPendingJobInMemory() store.PendingJob {
	return &pendingJobInMemory{
		indexByJobUUID:        make(map[string]*entities.Job),   // JobUUID => Job
		indexByTxHash:         make(map[string]*entities.Job),   // TxHash => Job
		aggregatedByChainUUID: make(map[string]map[string]bool), // ChainUUID => JobUUID => bool
		mux:                   &sync.RWMutex{},
	}
}

func (m *pendingJobInMemory) Add(_ context.Context, job *entities.Job) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexByJobUUID[job.UUID]; ok {
		return errors.AlreadyExistsError("job %q is duplicated", job.UUID)
	}

	m.indexByJobUUID[job.UUID] = job
	m.indexByTxHash[job.Transaction.Hash.String()] = job
	if _, ok := m.aggregatedByChainUUID[job.ChainUUID]; !ok {
		m.aggregatedByChainUUID[job.ChainUUID] = make(map[string]bool)
	}
	m.aggregatedByChainUUID[job.ChainUUID][job.UUID] = true
	return nil
}

func (m *pendingJobInMemory) Update(_ context.Context, job *entities.Job) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexByJobUUID[job.UUID]; !ok {
		return errors.NotFoundError("job %q is duplicated", job.UUID)
	}

	previousJob := m.indexByJobUUID[job.UUID]
	m.indexByJobUUID[job.UUID] = job
	delete(m.indexByTxHash, previousJob.Transaction.Hash.String())
	m.indexByTxHash[job.Transaction.Hash.String()] = job
	return nil
}

func (m *pendingJobInMemory) Remove(_ context.Context, jobUUID string) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexByJobUUID[jobUUID]; !ok {
		return errors.NotFoundError("job %q is not found", jobUUID)
	}

	job := m.indexByJobUUID[jobUUID]
	delete(m.aggregatedByChainUUID[job.ChainUUID], job.UUID)
	delete(m.indexByTxHash, job.Transaction.Hash.String())
	delete(m.indexByJobUUID, job.UUID)
	return nil
}

func (m *pendingJobInMemory) GetByTxHash(_ context.Context, chainUUID string, txHash *common.Hash) (*entities.Job, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	if job, ok := m.indexByTxHash[txHash.String()]; ok {
		if chainUUID == "" || job.ChainUUID == chainUUID {
			return job, nil
		}
	}

	return nil, errors.NotFoundError("job with hash %q is not found", txHash.String())
}

func (m *pendingJobInMemory) GetJobUUID(_ context.Context, jobUUID string) (*entities.Job, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	if _, ok := m.indexByJobUUID[jobUUID]; !ok {
		return nil, errors.NotFoundError("job %q is not found", jobUUID)
	}
	return m.indexByJobUUID[jobUUID], nil
}

func (m *pendingJobInMemory) ListPerChainUUID(_ context.Context, chainUUID string) ([]*entities.Job, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	jobs := []*entities.Job{}
	if _, ok := m.aggregatedByChainUUID[chainUUID]; !ok {
		return jobs, errors.NotFoundError("there is not jobs in chain %q", chainUUID)
	}

	for jobUUID := range m.aggregatedByChainUUID[chainUUID] {
		jobs = append(jobs, m.indexByJobUUID[jobUUID])
	}

	return jobs, nil
}

func (m *pendingJobInMemory) DeletePerChainUUID(ctx context.Context, chainUUID string) error {
	// @TODO Mutex
	if _, ok := m.aggregatedByChainUUID[chainUUID]; !ok {
		return errors.NotFoundError("there is not jobs in chain %q", chainUUID)
	}

	for jobUUID := range m.aggregatedByChainUUID[chainUUID] {
		_ = m.Remove(ctx, jobUUID)
	}
	return nil
}
