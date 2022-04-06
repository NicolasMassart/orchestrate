package builder

import (
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/transactions"
	"github.com/consensys/orchestrate/src/api/store"
)

type transactionUseCases struct {
	sendContract usecases.SendContractTxUseCase
	sendDeploy   usecases.SendDeployTxUseCase
	send         usecases.SendTxUseCase
	get          usecases.GetTxUseCase
	search       usecases.SearchTransactionsUseCase
	speedUp      usecases.SpeedUpTxUseCase
	callOff      usecases.CallOffTxUseCase
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
	sendTxUC := transactions.NewSendTxUseCase(db, searchChainsUC, jobUCs.Start(), jobUCs.Create(), getTransactionUC,
		getFaucetCandidateUC)

	return &transactionUseCases{
		sendContract: transactions.NewSendContractTxUseCase(sendTxUC, getContractUC),
		sendDeploy:   transactions.NewSendDeployTxUseCase(sendTxUC, getContractUC),
		send:         sendTxUC,
		get:          getTransactionUC,
		search:       transactions.NewSearchTransactionsUseCase(db, getTransactionUC),
		speedUp:      transactions.NewSpeedUpTxUseCase(getTransactionUC, jobUCs.retryTx),
		callOff:      transactions.NewCallOffTxUseCase(getTransactionUC, jobUCs.retryTx),
	}
}

func (u *transactionUseCases) SendContract() usecases.SendContractTxUseCase {
	return u.sendContract
}

func (u *transactionUseCases) SendDeploy() usecases.SendDeployTxUseCase {
	return u.sendDeploy
}

func (u *transactionUseCases) Send() usecases.SendTxUseCase {
	return u.send
}

func (u *transactionUseCases) Get() usecases.GetTxUseCase {
	return u.get
}

func (u *transactionUseCases) Search() usecases.SearchTransactionsUseCase {
	return u.search
}

func (u *transactionUseCases) SpeedUp() usecases.SpeedUpTxUseCase {
	return u.speedUp
}

func (u *transactionUseCases) CallOff() usecases.CallOffTxUseCase {
	return u.callOff
}
