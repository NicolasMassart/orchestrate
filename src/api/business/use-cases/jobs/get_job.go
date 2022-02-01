package jobs

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
)

const getJobComponent = "use-cases.get-job"

// getJobUseCase is a use case to get a job
type getJobUseCase struct {
	db     store.DB
	logger *log.Logger
}

// NewGetJobUseCase creates a new GetJobUseCase
func NewGetJobUseCase(db store.DB) usecases.GetJobUseCase {
	return &getJobUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(getJobComponent),
	}
}

// Execute gets a job
func (uc *getJobUseCase) Execute(ctx context.Context, jobUUID string, userInfo *multitenancy.UserInfo) (*entities.Job, error) {
	ctx = log.WithFields(ctx, log.Field("job", jobUUID))
	job, err := uc.db.Job().FindOneByUUID(ctx, jobUUID, userInfo.AllowedTenants, userInfo.Username, true)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(getJobComponent)
	}

	uc.logger.WithContext(ctx).Debug("job found successfully")
	return job, nil
}
