package dataagents

import (
	"github.com/consensys/orchestrate/src/api/store"
	pg "github.com/consensys/orchestrate/src/infra/database/postgres"
)

type PGAgents struct {
	tx            store.TransactionAgent
	job           store.JobAgent
	log           store.LogAgent
	schedule      store.ScheduleAgent
	txRequest     store.TransactionRequestAgent
	account       store.AccountAgent
	faucet        store.FaucetAgent
	contractEvent store.ContractEventAgent
	contract      store.ContractAgent
	chain         store.ChainAgent
}

func New(db pg.DB) *PGAgents {
	return &PGAgents{
		tx:            NewPGTransaction(db),
		job:           NewPGJob(db),
		log:           NewPGLog(db),
		schedule:      NewPGSchedule(db),
		txRequest:     NewPGTransactionRequest(db),
		account:       NewPGAccount(db),
		faucet:        NewPGFaucet(db),
		contractEvent: NewPGContractEvent(db),
		contract:      NewPGContract(db),
		chain:         NewPGChain(db),
	}
}

func (a *PGAgents) Job() store.JobAgent {
	return a.job
}

func (a *PGAgents) Log() store.LogAgent {
	return a.log
}

func (a *PGAgents) Schedule() store.ScheduleAgent {
	return a.schedule
}

func (a *PGAgents) Transaction() store.TransactionAgent {
	return a.tx
}

func (a *PGAgents) TransactionRequest() store.TransactionRequestAgent {
	return a.txRequest
}

func (a *PGAgents) Account() store.AccountAgent {
	return a.account
}

func (a *PGAgents) Faucet() store.FaucetAgent {
	return a.faucet
}

func (a *PGAgents) ContractEvent() store.ContractEventAgent {
	return a.contractEvent
}

func (a *PGAgents) Contract() store.ContractAgent {
	return a.contract
}

func (a *PGAgents) Chain() store.ChainAgent {
	return a.chain
}
