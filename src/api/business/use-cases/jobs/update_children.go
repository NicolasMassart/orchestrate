package jobs

import (
	"context"
	"fmt"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"

	"github.com/consensys/orchestrate/pkg/errors"

	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
)

const updateChildrenComponent = "use-cases.update-children"

// createJobUseCase is a use case to create a new transaction job
type updateChildrenUseCase struct {
	db     store.DB
	logger *log.Logger
}

// NewUpdateChildrenUseCase creates a new UpdateChildrenUseCase
func NewUpdateChildrenUseCase(db store.DB) usecases.UpdateChildrenUseCase {
	return &updateChildrenUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(updateChildrenComponent),
	}
}

func (uc updateChildrenUseCase) WithDBTransaction(dbtx store.DB) usecases.UpdateChildrenUseCase {
	uc.db = dbtx
	return &uc
}

func (uc *updateChildrenUseCase) Execute(ctx context.Context, jobUUID, parentJobUUID string,
	nextStatus entities.JobStatus, userInfo *multitenancy.UserInfo) error {
	ctx = log.WithFields(ctx, log.Field("job", jobUUID), log.Field("parent_job", parentJobUUID),
		log.Field("next_status", nextStatus))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("updating sibling and/or parent jobs")

	if !entities.IsFinalJobStatus(nextStatus) {
		errMsg := "expected final job status"
		err := errors.InvalidParameterError(errMsg)
		logger.WithError(err).Error("failed to update children jobs")
		return err
	}

	jobsToUpdate, err := uc.db.Job().Search(ctx, &entities.JobFilters{
		ParentJobUUID: parentJobUUID,
		Status:        entities.StatusPending,
	}, userInfo.AllowedTenants, userInfo.Username)

	if err != nil {
		return errors.FromError(err).ExtendComponent(updateChildrenComponent)
	}

	for _, job := range jobsToUpdate {
		// Skip mined job which trigger the update of sibling/children
		if job.UUID == jobUUID {
			continue
		}

		job.Status = nextStatus
		if err = uc.db.Job().Update(ctx, job); err != nil {
			return errors.FromError(err).ExtendComponent(updateChildrenComponent)
		}

		jobLog := &entities.Log{
			Status:  nextStatus,
			Message: fmt.Sprintf("sibling (or parent) job %s was mined instead", jobUUID),
		}
		_, err = uc.db.Log().Insert(ctx, jobLog, job.UUID)
		if err != nil {
			return errors.FromError(err).ExtendComponent(updateChildrenComponent)
		}

		logger.WithField("job", job.UUID).
			WithField("status", nextStatus).Debug("updated children/sibling job successfully")
	}

	logger.WithField("status", nextStatus).Info("children (and/or parent) jobs updated successfully")
	return nil
}
