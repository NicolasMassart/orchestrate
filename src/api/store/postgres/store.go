package postgres

import (
	"context"

	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/infra/postgres"
)

type PGStore struct {
	job           store.JobAgent
	schedule      store.ScheduleAgent
	txRequest     store.TransactionRequestAgent
	account       store.AccountAgent
	faucet        store.FaucetAgent
	contractEvent store.ContractEventAgent
	contract      store.ContractAgent
	chain         store.ChainAgent
	client        postgres.Client
}

var _ store.DB = &PGStore{}

func New(client postgres.Client) *PGStore {
	return &PGStore{
		job:           NewPGJob(client),
		schedule:      NewPGSchedule(client),
		txRequest:     NewPGTransactionRequest(client),
		account:       NewPGAccount(client),
		faucet:        NewPGFaucet(client),
		contractEvent: NewPGContractEvent(client),
		contract:      NewPGContract(client),
		chain:         NewPGChain(client),
		client:        client,
	}
}

func (s *PGStore) Job() store.JobAgent {
	return s.job
}

func (s *PGStore) Schedule() store.ScheduleAgent {
	return s.schedule
}

func (s *PGStore) TransactionRequest() store.TransactionRequestAgent {
	return s.txRequest
}

func (s *PGStore) Account() store.AccountAgent {
	return s.account
}

func (s *PGStore) Faucet() store.FaucetAgent {
	return s.faucet
}

func (s *PGStore) ContractEvent() store.ContractEventAgent {
	return s.contractEvent
}

func (s *PGStore) Contract() store.ContractAgent {
	return s.contract
}

func (s *PGStore) Chain() store.ChainAgent {
	return s.chain
}

func (s *PGStore) RunInTransaction(ctx context.Context, persist func(a store.DB) error) error {
	return s.client.RunInTransaction(ctx, func(dbTx postgres.Client) error {
		return persist(New(dbTx))
	})
}
