package builder

import (
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/contracts"
	"github.com/consensys/orchestrate/src/api/store"
)

type contractUseCases struct {
	GetContractsCatalogUC      usecases.GetContractsCatalogUseCase
	getContractEvents          usecases.GetContractEventsUseCase
	getContractTags            usecases.GetContractTagsUseCase
	registerContractDeployment usecases.RegisterContractDeploymentUseCase
	registerContractUC         usecases.RegisterContractUseCase
	getContractUC              usecases.GetContractUseCase
	searchContractUC           usecases.SearchContractUseCase
}

func newContractUseCases(db store.DB) *contractUseCases {
	getContractUC := contracts.NewGetContractUseCase(db.Contract())

	return &contractUseCases{
		registerContractUC:         contracts.NewRegisterContractUseCase(db),
		getContractUC:              getContractUC,
		GetContractsCatalogUC:      contracts.NewGetCatalogUseCase(db.Contract()),
		getContractEvents:          contracts.NewGetEventsUseCase(db.ContractEvent()),
		getContractTags:            contracts.NewGetTagsUseCase(db.Contract()),
		registerContractDeployment: contracts.NewRegisterDeploymentUseCase(db.Contract()),
		searchContractUC:           contracts.NewSearchContractUseCase(db.Contract()),
	}
}

func (u *contractUseCases) GetContract() usecases.GetContractUseCase {
	return u.getContractUC
}

func (u *contractUseCases) RegisterContract() usecases.RegisterContractUseCase {
	return u.registerContractUC
}

func (u *contractUseCases) GetContractsCatalog() usecases.GetContractsCatalogUseCase {
	return u.GetContractsCatalogUC
}

func (u *contractUseCases) GetContractEvents() usecases.GetContractEventsUseCase {
	return u.getContractEvents
}

func (u *contractUseCases) GetContractTags() usecases.GetContractTagsUseCase {
	return u.getContractTags
}

func (u *contractUseCases) SetContractCodeHash() usecases.RegisterContractDeploymentUseCase {
	return u.registerContractDeployment
}

func (u *contractUseCases) SearchContract() usecases.SearchContractUseCase {
	return u.searchContractUC
}
