package builder

import (
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/accounts"
	"github.com/consensys/orchestrate/src/api/store"
	qkmclient "github.com/consensys/quorum-key-manager/pkg/client"
)

type accountUseCases struct {
	create usecases.CreateAccountUseCase
	get    usecases.GetAccountUseCase
	search usecases.SearchAccountsUseCase
	update usecases.UpdateAccountUseCase
}

func newAccountUseCases(
	db store.DB,
	keyManagerClient qkmclient.EthClient,
	searchChainsUC usecases.SearchChainsUseCase,
	sendTxUC usecases.SendTxUseCase,
	getFaucetCandidateUC usecases.GetFaucetCandidateUseCase,
) *accountUseCases {
	searchAccountsUC := accounts.NewSearchAccountsUseCase(db)
	fundAccountUC := accounts.NewFundAccountUseCase(searchChainsUC, sendTxUC, getFaucetCandidateUC)

	return &accountUseCases{
		create: accounts.NewCreateAccountUseCase(db, searchAccountsUC, fundAccountUC, keyManagerClient),
		get:    accounts.NewGetAccountUseCase(db),
		search: searchAccountsUC,
		update: accounts.NewUpdateAccountUseCase(db),
	}
}

func (u *accountUseCases) Get() usecases.GetAccountUseCase {
	return u.get
}

func (u *accountUseCases) Search() usecases.SearchAccountsUseCase {
	return u.search
}

func (u *accountUseCases) Create() usecases.CreateAccountUseCase {
	return u.create
}

func (u *accountUseCases) Update() usecases.UpdateAccountUseCase {
	return u.update
}
