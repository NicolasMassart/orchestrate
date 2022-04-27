package usecases

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=jobs.go -destination=mocks/jobs.go -package=mocks

type PendingJob interface {
	Execute(ctx context.Context, job *entities.Job) error
}

type FailedJob interface {
	Execute(ctx context.Context, job *entities.Job, errMsg string) error
}

type MinedJob interface {
	Execute(ctx context.Context, job *entities.Job) error
}

type RetryJob interface {
	Execute(ctx context.Context, job *entities.Job, lastChildUUID string, nChildren int) (string, error)
}

type JobUseCases interface {
	PendingJobUseCase() PendingJob
	FailedJobUseCase() FailedJob
	MinedJobUseCase() MinedJob
	RetryJobUseCase() RetryJob
}
