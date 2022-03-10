package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/metrics"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
)

const updateJobComponent = "use-cases.update-job"

type updateJobUseCase struct {
	db                    store.DB
	updateChildrenUseCase usecases.UpdateChildrenUseCase
	startNextJobUseCase   usecases.StartNextJobUseCase
	metrics               metrics.TransactionSchedulerMetrics
	logger                *log.Logger
}

func NewUpdateJobUseCase(db store.DB, updateChildrenUseCase usecases.UpdateChildrenUseCase,
	startJobUC usecases.StartNextJobUseCase, m metrics.TransactionSchedulerMetrics) usecases.UpdateJobUseCase {
	return &updateJobUseCase{
		db:                    db,
		updateChildrenUseCase: updateChildrenUseCase,
		startNextJobUseCase:   startJobUC,
		metrics:               m,
		logger:                log.NewLogger().SetComponent(updateJobComponent),
	}
}

func (uc *updateJobUseCase) Execute(ctx context.Context, nextJob *entities.Job, nextStatus entities.JobStatus,
	logMessage string, userInfo *multitenancy.UserInfo) (*entities.Job, error) {
	ctx = log.WithFields(ctx, log.Field("job", nextJob.UUID), log.Field("next_status", nextStatus))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("updating job")

	job, err := uc.db.Job().FindOneByUUID(ctx, nextJob.UUID, userInfo.AllowedTenants, userInfo.Username, true)
	if err != nil {
		return nil, err
	}

	if entities.IsFinalJobStatus(job.Status) {
		errMessage := fmt.Sprintf("job status %s is final, cannot be updated", job.Status)
		logger.WithField("status", job.Status).Error(errMessage)
		return nil, errors.InvalidParameterError(errMessage).ExtendComponent(updateJobComponent)
	}

	if nextJob.Transaction != nil {
		nextJob.Transaction, err = uc.db.Transaction().Update(ctx, nextJob.Transaction, nextJob.UUID)
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(updateJobComponent)
		}
	}

	if len(nextJob.Labels) > 0 {
		job.Labels = nextJob.Labels
	}
	if nextJob.InternalData != nil {
		job.InternalData = nextJob.InternalData
	}

	var jobLog *entities.Log
	// We are not forced to update the status
	if nextStatus != "" && !canUpdateStatus(nextStatus, job.Status) {
		errMessage := "invalid status update for the current job state"
		logger.WithField("status", job.Status).WithField("next_status", nextStatus).Error(errMessage)
		return nil, errors.InvalidStateError(errMessage).ExtendComponent(updateJobComponent)
	} else if nextStatus != "" {
		jobLog = &entities.Log{
			Status:  nextStatus,
			Message: logMessage,
		}
	}

	// In case of status update
	err = uc.updateJob(ctx, job, jobLog, userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateJobComponent)
	}

	if (nextStatus == entities.StatusMined || nextStatus == entities.StatusStored) && job.NextJobUUID != "" {
		err = uc.startNextJobUseCase.Execute(ctx, job.UUID, userInfo)
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(updateJobComponent)
		}
	}

	logger.Info("job updated successfully")
	return job, nil
}

func (uc *updateJobUseCase) updateJob(ctx context.Context, job *entities.Job, jobLog *entities.Log, userInfo *multitenancy.UserInfo) error {
	logger := uc.logger.WithContext(ctx)

	// Does current job belong to a parent/children chains?
	var parentJobUUID string
	if job.InternalData.ParentJobUUID != "" {
		parentJobUUID = job.InternalData.ParentJobUUID
	} else if job.InternalData.RetryInterval != 0 {
		parentJobUUID = job.UUID
	}

	prevLogModel := job.Logs[len(job.Logs)-1]
	err := uc.db.RunInTransaction(ctx, func(dbtx store.DB) error {
		// We should lock ONLY when there is children jobs
		if parentJobUUID != "" {
			logger.WithField("parent_job", parentJobUUID).Debug("lock parent job row for update")
			if err := dbtx.Job().LockOneByUUID(ctx, parentJobUUID); err != nil {
				return err
			}

			// Refresh jobModel after lock to ensure nothing was updated
			refreshedJobModel, err := uc.db.Job().FindOneByUUID(ctx, job.UUID,
				userInfo.AllowedTenants, userInfo.Username, false)
			if err != nil {
				return err
			}

			if refreshedJobModel.UpdatedAt != job.UpdatedAt {
				errMessage := "job status was updated since user request was sent"
				logger.WithField("status", job.Status).Error(errMessage)
				return errors.InvalidStateError(errMessage).ExtendComponent(updateJobComponent)
			}
		}

		if jobLog != nil {
			var err error
			jobLog, err = dbtx.Log().Insert(ctx, jobLog, job.UUID)
			if err != nil {
				return err
			}

			job.Logs = append(job.Logs, jobLog)
			if updateNextJobStatus(prevLogModel.Status, jobLog.Status) {
				job.Status = jobLog.Status
			}
		}

		if err := dbtx.Job().Update(ctx, job); err != nil {
			return err
		}

		// if we updated to MINED, we need to update the children and sibling jobs to NEVER_MINED
		if parentJobUUID != "" && jobLog != nil && jobLog.Status == entities.StatusMined {
			der := uc.updateChildrenUseCase.WithDBTransaction(dbtx).Execute(ctx, job.UUID, parentJobUUID, entities.StatusNeverMined, userInfo)
			if der != nil {
				return der
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Metrics observe request latency over job status changes
	if jobLog != nil {
		uc.addMetrics(job.UpdatedAt.Sub(prevLogModel.CreatedAt), prevLogModel.Status, jobLog.Status, job.ChainUUID)
	}

	return nil
}

func updateNextJobStatus(prevStatus, nextStatus entities.JobStatus) bool {
	if nextStatus == entities.StatusResending {
		return false
	}
	if nextStatus == entities.StatusWarning {
		return false
	}
	if nextStatus == entities.StatusFailed && prevStatus == entities.StatusResending {
		return false
	}

	return true
}

func canUpdateStatus(nextStatus, status entities.JobStatus) bool {
	switch nextStatus {
	case entities.StatusCreated:
		return false
	case entities.StatusStarted:
		return status == entities.StatusCreated
	case entities.StatusPending:
		return status == entities.StatusStarted || status == entities.StatusRecovering
	case entities.StatusResending:
		return status == entities.StatusPending || status == entities.StatusResending
	case entities.StatusRecovering:
		return status == entities.StatusStarted || status == entities.StatusRecovering || status == entities.StatusPending
	case entities.StatusMined, entities.StatusNeverMined:
		return status == entities.StatusPending
	case entities.StatusStored:
		return status == entities.StatusStarted || status == entities.StatusRecovering
	case entities.StatusFailed:
		return status == entities.StatusStarted || status == entities.StatusRecovering || status == entities.StatusPending || status == entities.StatusWarning || status == entities.StatusResending
	default: // For warning, they can be added at any time
		return true
	}
}

func (uc *updateJobUseCase) addMetrics(elapseTime time.Duration, previousStatus, nextStatus entities.JobStatus, chainUUID string) {
	if previousStatus == nextStatus {
		return
	}

	baseLabels := []string{
		"chain_uuid", chainUUID,
	}

	switch nextStatus {
	case entities.StatusMined:
		uc.metrics.MinedLatencyHistogram().With(append(baseLabels,
			"prev_status", string(previousStatus),
			"status", string(nextStatus),
		)...).Observe(elapseTime.Seconds())
	default:
		uc.metrics.JobsLatencyHistogram().With(append(baseLabels,
			"prev_status", string(previousStatus),
			"status", string(nextStatus),
		)...).Observe(elapseTime.Seconds())
	}
}
