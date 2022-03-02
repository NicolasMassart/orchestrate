package usecases

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=tx-listener.go -destination=mocks/tx-listener.go -package=mocks

type NotifyMinedJob interface {
	Execute(ctx context.Context, job *entities.Job) error
}

type SendNotification interface {
	Execute(ctx context.Context, job *entities.Job) error
}

type RegisterDeployedContract interface {
	Execute(ctx context.Context, job *entities.Job) error
}

type UpdateChainHead interface {
	Execute(ctx context.Context, chainUUID string, nextChainHead uint64) error
}
