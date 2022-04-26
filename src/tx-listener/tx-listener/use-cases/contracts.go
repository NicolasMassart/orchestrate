package usecases

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=contracts.go -destination=mocks/contracts.go -package=mocks

type RegisterDeployedContract interface {
	Execute(ctx context.Context, job *entities.Job) error
}

type ContractsUseCases interface {
	RegisterDeployedContractUseCase() RegisterDeployedContract
}
