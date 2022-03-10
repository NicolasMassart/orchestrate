package usecases

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=sender.go -destination=mocks/sender.go -package=mocks

type SendETHRawTxUseCase interface {
	Execute(ctx context.Context, job *entities.Job) error
}

type SendETHTxUseCase interface {
	Execute(ctx context.Context, job *entities.Job) error
}

type SendEEAPrivateTxUseCase interface {
	Execute(ctx context.Context, job *entities.Job) error
}

type SendGoQuorumPrivateTxUseCase interface {
	Execute(ctx context.Context, job *entities.Job) error
}

type SendGoQuorumMarkingTxUseCase interface {
	Execute(ctx context.Context, job *entities.Job) error
}
