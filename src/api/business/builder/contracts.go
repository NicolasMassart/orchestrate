package builder

import (
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/contracts"
	"github.com/consensys/orchestrate/src/api/store"
)

type contractUseCases struct {
	getContractsCatalogUC      usecases.GetContractsCatalogUseCase
	getContractEvents          usecases.GetContractEventsUseCase
	getContractTags            usecases.GetContractTagsUseCase
	registerContractDeployment usecases.RegisterContractDeploymentUseCase
	registerContractUC         usecases.RegisterContractUseCase
	getContractUC              usecases.GetContractUseCase
	searchContractUC           usecases.SearchContractUseCase
	decodeLogUC                usecases.DecodeEventLogUseCase
}

func newContractUseCases(db store.DB) *contractUseCases {
	getContractUC := contracts.NewGetContractUseCase(db.Contract())
	getContractEventUC := contracts.NewGetEventsUseCase(db.ContractEvent())
	
	return &contractUseCases{
		registerContractUC:         contracts.NewRegisterContractUseCase(db),
		getContractUC:              getContractUC,
		getContractsCatalogUC:      contracts.NewGetCatalogUseCase(db.Contract()),
		decodeLogUC:                contracts.NewDecodeEventLogUseCase(getContractEventUC),
		getContractEvents:          getContractEventUC,
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
	return u.getContractsCatalogUC
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

func (u *contractUseCases) DecodeLog() usecases.DecodeEventLogUseCase {
	return u.decodeLogUC
}
