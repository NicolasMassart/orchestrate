package builder

import (
	"github.com/consensys/orchestrate/src/tx-listener/store"
	in_memory "github.com/consensys/orchestrate/src/tx-listener/store/in-memory"
)

type storeState struct {
	chain           store.Chain
	pendingJob      store.PendingJob
	retryJobSession store.RetryJobSession
}

func (s *storeState) ChainState() store.Chain {
	return s.chain
}

func (s *storeState) PendingJobState() store.PendingJob {
	return s.pendingJob
}

func (s *storeState) RetryJobSessionState() store.RetryJobSession {
	return s.retryJobSession
}

func NewStoreState() store.State {
	chainInMemory := in_memory.NewChainInMemory()
	pendingJobInMemory := in_memory.NewPendingJobInMemory()
	retryJobSessionInMemory := in_memory.NewRetryJobSessionInMemory()
	return &storeState{
		chain:           chainInMemory,
		pendingJob:      pendingJobInMemory,
		retryJobSession: retryJobSessionInMemory,
	}
}
