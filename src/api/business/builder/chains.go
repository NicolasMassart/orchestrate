package builder

import (
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/chains"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/infra/ethclient"
)

type chainUseCases struct {
	register usecases.RegisterChainUseCase
	update   usecases.UpdateChainUseCase
	get      usecases.GetChainUseCase
	search   usecases.SearchChainsUseCase
	delete   usecases.DeleteChainUseCase
}

func newChainUseCases(db store.DB, ec ethclient.Client) *chainUseCases {
	searchChainsUC := chains.NewSearchChainsUseCase(db)
	getChainUC := chains.NewGetChainUseCase(db)

	return &chainUseCases{
		register: chains.NewRegisterChainUseCase(db, searchChainsUC, ec),
		update:   chains.NewUpdateChainUseCase(db, getChainUC),
		get:      getChainUC,
		search:   searchChainsUC,
		delete:   chains.NewDeleteChainUseCase(db, getChainUC),
	}
}

func (u *chainUseCases) Register() usecases.RegisterChainUseCase {
	return u.register
}

func (u *chainUseCases) Update() usecases.UpdateChainUseCase {
	return u.update
}

func (u *chainUseCases) Get() usecases.GetChainUseCase {
	return u.get
}

func (u *chainUseCases) Search() usecases.SearchChainsUseCase {
	return u.search
}

func (u *chainUseCases) Delete() usecases.DeleteChainUseCase {
	return u.delete
}
