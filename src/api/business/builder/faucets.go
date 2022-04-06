package builder

import (
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/faucets"
	"github.com/consensys/orchestrate/src/api/store"
)

type faucetUseCases struct {
	register usecases.RegisterFaucetUseCase
	update   usecases.UpdateFaucetUseCase
	get      usecases.GetFaucetUseCase
	search   usecases.SearchFaucetsUseCase
	delete   usecases.DeleteFaucetUseCase
}

func newFaucetUseCases(db store.DB) *faucetUseCases {
	search := faucets.NewSearchFaucets(db)

	return &faucetUseCases{
		register: faucets.NewRegisterFaucetUseCase(db, search),
		update:   faucets.NewUpdateFaucetUseCase(db),
		get:      faucets.NewGetFaucetUseCase(db),
		search:   search,
		delete:   faucets.NewDeleteFaucetUseCase(db),
	}
}

func (u *faucetUseCases) Register() usecases.RegisterFaucetUseCase {
	return u.register
}

func (u *faucetUseCases) Update() usecases.UpdateFaucetUseCase {
	return u.update
}

func (u *faucetUseCases) Get() usecases.GetFaucetUseCase {
	return u.get
}

func (u *faucetUseCases) Search() usecases.SearchFaucetsUseCase {
	return u.search
}

func (u *faucetUseCases) Delete() usecases.DeleteFaucetUseCase {
	return u.delete
}
