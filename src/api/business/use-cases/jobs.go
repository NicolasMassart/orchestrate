package usecases

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

//go:generate mockgen -source=jobs.go -destination=mocks/jobs.go -package=mocks

type JobUseCases interface {
	Create() CreateJobUseCase
	Get() GetJobUseCase
	Start() StartJobUseCase
	ResendTx() ResendJobTxUseCase
	RetryTx() RetryJobTxUseCase
	Update() UpdateJobUseCase
	Search() SearchJobsUseCase
}

type CreateJobUseCase interface {
	Execute(ctx context.Context, job *entities.Job, userInfo *multitenancy.UserInfo) (*entities.Job, error)
}

type GetJobUseCase interface {
	Execute(ctx context.Context, jobUUID string, userInfo *multitenancy.UserInfo) (*entities.Job, error)
}

type SearchJobsUseCase interface {
	Execute(ctx context.Context, filters *entities.JobFilters, userInfo *multitenancy.UserInfo) ([]*entities.Job, error)
}

type StartJobUseCase interface {
	Execute(ctx context.Context, jobUUID string, userInfo *multitenancy.UserInfo) error
}

type StartNextJobUseCase interface {
	Execute(ctx context.Context, prevJobUUID string, userInfo *multitenancy.UserInfo) error
}

type UpdateJobUseCase interface {
	Execute(ctx context.Context, jobEntity *entities.Job, nextStatus entities.JobStatus, logMessage string, userInfo *multitenancy.UserInfo) (*entities.Job, error)
}

type UpdateJobStatusUseCase interface {
	Execute(ctx context.Context, job *entities.Job, nextStatus entities.JobStatus, msg string) error
}

type ResendJobTxUseCase interface {
	Execute(ctx context.Context, jobUUID string, userInfo *multitenancy.UserInfo) error
}

type RetryJobTxUseCase interface {
	Execute(ctx context.Context, jobUUID string, gasIncrement float64, data hexutil.Bytes, userInfo *multitenancy.UserInfo) error
}
