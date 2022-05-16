package builder

import (
	"github.com/consensys/orchestrate/src/tx-listener/store"
	in_memory "github.com/consensys/orchestrate/src/tx-listener/store/in-memory"
)

type storeState struct {
	chain           store.Chain
	pendingJob      store.PendingJob
	retryJobSession store.RetryJobSession
	subscriptions   store.Subscriptions
	messenger       store.Message
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

func (s *storeState) SubscriptionState() store.Subscriptions {
	return s.subscriptions
}

func (s *storeState) MessengerState() store.Message {
	return s.messenger
}

func NewStoreState() store.State {
	chainInMemory := in_memory.NewChainState()
	pendingJobInMemory := in_memory.NewPendingJobState()
	retryJobSessionInMemory := in_memory.NewRetryJobSessionState()
	subscriptionInMemory := in_memory.NewSubscriptionState()
	messengerInMemory := in_memory.NewMessengerState()

	return &storeState{
		chain:           chainInMemory,
		pendingJob:      pendingJobInMemory,
		retryJobSession: retryJobSessionInMemory,
		subscriptions:   subscriptionInMemory,
		messenger:       messengerInMemory,
	}
}
