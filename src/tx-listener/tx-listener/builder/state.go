package builder

import (
	"github.com/consensys/orchestrate/src/tx-listener/store"
	in_memory "github.com/consensys/orchestrate/src/tx-listener/store/in-memory"
)

type storeState struct {
	chain         store.Chain
	pendingJob    store.PendingJob
	retrySessions store.RetrySessions
}

func (s *storeState) ChainState() store.Chain {
	return s.chain
}

func (s *storeState) PendingJobState() store.PendingJob {
	return s.pendingJob
}

func (s *storeState) RetrySessionsState() store.RetrySessions {
	return s.retrySessions
}

func NewStoreState() store.State {
	chainInMemory := in_memory.NewChainInMemory()
	pendingJobInMemory := in_memory.NewPendingJobInMemory()
	retrySessionInMemory := in_memory.NewRetrySessionInMemory()
	return &storeState{
		chain:         chainInMemory,
		pendingJob:    pendingJobInMemory,
		retrySessions: retrySessionInMemory,
	}
}
