package builder

import (
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/transactions"
	"github.com/consensys/orchestrate/src/api/store"
)

type transactionUseCases struct {
	sendContractTransaction usecases.SendContractTxUseCase
	sendDeployTransaction   usecases.SendDeployTxUseCase
	sendTransaction         usecases.SendTxUseCase
	getTransaction          usecases.GetTxUseCase
	searchTransactions      usecases.SearchTransactionsUseCase
	speedUpTransactions     usecases.SpeedUpTxUseCase
	callOffTransactions     usecases.CallOffTxUseCase
}

func newTransactionUseCases(
	db store.DB,
	searchChainsUC usecases.SearchChainsUseCase,
	getFaucetCandidateUC usecases.GetFaucetCandidateUseCase,
	schedulesUCs *scheduleUseCases,
	jobUCs *jobUseCases,
	getContractUC usecases.GetContractUseCase,
) *transactionUseCases {
	getTransactionUC := transactions.NewGetTxUseCase(db, schedulesUCs.GetSchedule())
	sendTxUC := transactions.NewSendTxUseCase(db, searchChainsUC, jobUCs.StartJob(), jobUCs.CreateJob(), getTransactionUC,
		getFaucetCandidateUC)

	return &transactionUseCases{
		sendContractTransaction: transactions.NewSendContractTxUseCase(sendTxUC, getContractUC),
		sendDeployTransaction:   transactions.NewSendDeployTxUseCase(sendTxUC, getContractUC),
		sendTransaction:         sendTxUC,
		getTransaction:          getTransactionUC,
		searchTransactions:      transactions.NewSearchTransactionsUseCase(db, getTransactionUC),
		speedUpTransactions:     transactions.NewSpeedUpTxUseCase(getTransactionUC, jobUCs.retryJobTx),
		callOffTransactions:     transactions.NewCallOffTxUseCase(getTransactionUC, jobUCs.retryJobTx),
	}
}

func (u *transactionUseCases) SendContractTransaction() usecases.SendContractTxUseCase {
	return u.sendContractTransaction
}

func (u *transactionUseCases) SendDeployTransaction() usecases.SendDeployTxUseCase {
	return u.sendDeployTransaction
}

func (u *transactionUseCases) SendTransaction() usecases.SendTxUseCase {
	return u.sendTransaction
}

func (u *transactionUseCases) GetTransaction() usecases.GetTxUseCase {
	return u.getTransaction
}

func (u *transactionUseCases) SearchTransactions() usecases.SearchTransactionsUseCase {
	return u.searchTransactions
}

func (u *transactionUseCases) SpeedUpTransaction() usecases.SpeedUpTxUseCase {
	return u.speedUpTransactions
}

func (u *transactionUseCases) CallOffTransaction() usecases.CallOffTxUseCase {
	return u.callOffTransactions
}
