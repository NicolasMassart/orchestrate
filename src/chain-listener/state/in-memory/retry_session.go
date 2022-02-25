package inmemory

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/chain-listener/state"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/ethereum/go-ethereum/common"
)

type retrySessionInMemory struct {
	indexBySessID         map[string]*entities.Job   // sessID => Job
	indexByTxHash         map[string]string          // TxHash => SessID
	aggregatedByChainUUID map[string]map[string]bool // ChainUUID => SessID => Bool
}

func NewRetrySessionInMemory() state.RetrySessions {
	return &retrySessionInMemory{
		indexBySessID:         make(map[string]*entities.Job),
		indexByTxHash:         make(map[string]string),
		aggregatedByChainUUID: make(map[string]map[string]bool),
	}
}

func (m *retrySessionInMemory) Add(_ context.Context, sessID string, job *entities.Job) error {
	// @TODO Mutex
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
	// @TODO Mutex
	if _, ok := m.indexBySessID[sessID]; !ok {
		return errors.NotFoundError("retry session %q is not found", sessID)
	}

	job := m.indexBySessID[sessID]
	delete(m.aggregatedByChainUUID[job.ChainUUID], sessID)
	delete(m.indexByTxHash, job.Transaction.Hash.String())
	delete(m.indexBySessID, sessID)
	return nil
}

func (m *retrySessionInMemory) SearchByTxHash(_ context.Context, txHash *common.Hash) (string, error) {
	// @TODO Mutex
	if _, ok := m.indexByTxHash[txHash.String()]; !ok {
		return "", errors.NotFoundError("retry session with hash %q is not found", txHash.String())
	}
	return m.indexByTxHash[txHash.String()], nil
}

func (m *retrySessionInMemory) ListByChainUUID(_ context.Context, chainUUID string) ([]string, error) {
	// @TODO Mutex
	sessionIDs := []string{}
	if _, ok := m.aggregatedByChainUUID[chainUUID]; !ok {
		return sessionIDs, errors.NotFoundError("there is not retry sessions in chain %q", chainUUID)
	}

	for jobUUID := range m.aggregatedByChainUUID[chainUUID] {
		sessionIDs = append(sessionIDs, m.indexByTxHash[jobUUID])
	}

	return sessionIDs, nil
}

func (m *retrySessionInMemory) DeletePerChainUUID(ctx context.Context, chainUUID string) error {
	// @TODO Mutex
	if _, ok := m.aggregatedByChainUUID[chainUUID]; !ok {
		return errors.NotFoundError("there is not retry sessions in chain %q", chainUUID)
	}

	for sessID := range m.aggregatedByChainUUID[chainUUID] {
		_ = m.Remove(ctx, sessID)
	}
	return nil
}
