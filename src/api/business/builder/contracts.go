package builder

import (
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/contracts"
	"github.com/consensys/orchestrate/src/api/store"
)

type contractUseCases struct {
	getCatalog         usecases.GetContractsCatalogUseCase
	getContractEvents  usecases.GetContractEventsUseCase
	getTags            usecases.GetContractTagsUseCase
	registerDeployment usecases.RegisterContractDeploymentUseCase
	register           usecases.RegisterContractUseCase
	get                usecases.GetContractUseCase
	search             usecases.SearchContractUseCase
	decodeLog          usecases.DecodeEventLogUseCase
}

var _ usecases.ContractUseCases = &contractUseCases{}

func newContractUseCases(db store.DB) *contractUseCases {
	getContractUC := contracts.NewGetContractUseCase(db.Contract())
	getContractEventUC := contracts.NewGetEventsUseCase(db.ContractEvent())

	return &contractUseCases{
		register:           contracts.NewRegisterContractUseCase(db),
		get:                getContractUC,
		getCatalog:         contracts.NewGetCatalogUseCase(db.Contract()),
		decodeLog:          contracts.NewDecodeEventLogUseCase(db, getContractEventUC),
		getContractEvents:  getContractEventUC,
		getTags:            contracts.NewGetTagsUseCase(db.Contract()),
		registerDeployment: contracts.NewRegisterDeploymentUseCase(db.Contract()),
		search:             contracts.NewSearchContractUseCase(db.Contract()),
	}
}

func (u *contractUseCases) Get() usecases.GetContractUseCase {
	return u.get
}

func (u *contractUseCases) Register() usecases.RegisterContractUseCase {
	return u.register
}

func (u *contractUseCases) GetCatalog() usecases.GetContractsCatalogUseCase {
	return u.getCatalog
}

func (u *contractUseCases) GetContractEvents() usecases.GetContractEventsUseCase {
	return u.getContractEvents
}

func (u *contractUseCases) GetTags() usecases.GetContractTagsUseCase {
	return u.getTags
}

func (u *contractUseCases) SetCodeHash() usecases.RegisterContractDeploymentUseCase {
	return u.registerDeployment
}

func (u *contractUseCases) Search() usecases.SearchContractUseCase {
	return u.search
}

func (u *contractUseCases) DecodeLog() usecases.DecodeEventLogUseCase {
	return u.decodeLog
}
