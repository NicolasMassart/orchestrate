package jobs

import (
	"context"

	log "github.com/sirupsen/logrus"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/errors"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/store"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/transaction-scheduler/entities"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/transaction-scheduler/parsers"
)

//go:generate mockgen -source=get_job.go -destination=mocks/get_job.go -package=mocks

const getJobComponent = "use-cases.get-job"

type GetJobUseCase interface {
	Execute(ctx context.Context, jobUUID, tenantID string) (*entities.Job, error)
}

// getJobUseCase is a use case to get a schedule
type getJobUseCase struct {
	db store.DB
}

// NewGetJobUseCase creates a new GetJobUseCase
func NewGetJobUseCase(db store.DB) GetJobUseCase {
	return &getJobUseCase{
		db: db,
	}
}

// Execute gets a schedule
func (uc *getJobUseCase) Execute(ctx context.Context, jobUUID, tenantID string) (*entities.Job, error) {
	log.WithContext(ctx).
		WithField("job_uuid", jobUUID).
		Debug("getting job")

	jobModel, err := uc.db.Job().FindOneByUUID(ctx, jobUUID, tenantID)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(getJobComponent)
	}

	log.WithContext(ctx).
		WithField("job_uuid", jobUUID).
		Info("job found successfully")

	return parsers.NewJobEntityFromModels(jobModel), nil
}