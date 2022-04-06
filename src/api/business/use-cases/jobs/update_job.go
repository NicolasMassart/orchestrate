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
	db             store.DB
	startNextJobUC usecases.StartNextJobUseCase
	notifyUC       usecases.NotifyTransactionUseCase
	metrics        metrics.TransactionSchedulerMetrics
	logger         *log.Logger
}

func NewUpdateJobUseCase(
	db store.DB,
	startNextJobUC usecases.StartNextJobUseCase,
	m metrics.TransactionSchedulerMetrics,
	notifyUC usecases.NotifyTransactionUseCase,
) usecases.UpdateJobUseCase {
	return &updateJobUseCase{
		db:             db,
		notifyUC:       notifyUC,
		startNextJobUC: startNextJobUC,
		metrics:        m,
		logger:         log.NewLogger().SetComponent(updateJobComponent),
	}
}

func (uc *updateJobUseCase) Execute(ctx context.Context, nextJob *entities.Job, nextStatus entities.JobStatus,
	nextStatusMsg string, userInfo *multitenancy.UserInfo) (*entities.Job, error) {
	ctx = log.WithFields(ctx, log.Field("job", nextJob.UUID), log.Field("next_status", nextStatus))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("updating job")

	prevJob, err := uc.db.Job().FindOneByUUID(ctx, nextJob.UUID, userInfo.AllowedTenants, userInfo.Username, false)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateJobComponent)
	}

	if nextStatus != "" && isValidJobStatus(nextStatus) {
		nextJob.Status = nextStatus
	}

	err = uc.updateJob(ctx, nextJob, nextStatus, nextStatusMsg, prevJob.InternalData.ParentJobUUID, userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateJobComponent)
	}

	uc.addJobStatusMetrics(prevJob, nextStatus)

	job, err := uc.db.Job().FindOneByUUID(ctx, nextJob.UUID, userInfo.AllowedTenants, userInfo.Username, true)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateJobComponent)
	}

	switch nextStatus {
	case entities.StatusMined:
		job.Receipt = nextJob.Receipt
		err = uc.notifyUC.Execute(ctx, job, "", userInfo)
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(updateJobComponent)
		}

		err = uc.startNextJobUC.Execute(ctx, job.UUID, userInfo)
	case entities.StatusFailed:
		err = uc.notifyUC.Execute(ctx, job, nextStatusMsg, userInfo)
	case entities.StatusStored:
		err = uc.startNextJobUC.Execute(ctx, job.UUID, userInfo)
	}
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateJobComponent)
	}

	return job, nil
}

func (uc *updateJobUseCase) updateJob(ctx context.Context, job *entities.Job, status entities.JobStatus,
	statusMsg, parentJobUUID string, userInfo *multitenancy.UserInfo) error {

	var jobLog *entities.Log
	if status != "" {
		jobLog = &entities.Log{
			Status:  status,
			Message: statusMsg,
		}
	}

	err := uc.db.RunInTransaction(ctx, func(dbtx store.DB) error {
		err := dbtx.Job().Update(ctx, job, jobLog)
		if err != nil {
			return err
		}

		// if we updated to MINED, we need to update the children and sibling jobs to NEVER_MINED
		if status != entities.StatusMined {
			return nil
		}

		if parentJobUUID == "" {
			parentJobUUID = job.UUID
		}

		siblingJobs, err := dbtx.Job().GetSiblingJobs(ctx, parentJobUUID, userInfo.AllowedTenants, userInfo.Username)
		if err != nil {
			return err
		}

		for _, siblingJob := range siblingJobs {
			// Skip mined job which trigger the update of sibling/children
			if job.UUID == siblingJob.UUID {
				continue
			}

			// Skip not pending sibling jobs
			if siblingJob.Status != entities.StatusPending {
				continue
			}

			siblingJob.Status = entities.StatusNeverMined
			err = dbtx.Job().Update(ctx, siblingJob, &entities.Log{
				Status:  entities.StatusNeverMined,
				Message: fmt.Sprintf("sibling (or parent) job %s was mined instead", job.UUID),
			})
			if err != nil {
				return errors.FromError(err).ExtendComponent(updateJobComponent)
			}
			uc.logger.WithField("job", siblingJob.UUID).
				WithField("status", entities.StatusNeverMined).
				Debug("updated job successfully")
		}
		return nil
	})

	if err != nil {
		uc.logger.WithError(err).Error("failed to update job")
		return errors.FromError(err).ExtendComponent(updateJobComponent)
	}

	uc.logger.WithField("job", job.UUID).WithField("status", status).Debug("updated job successfully")
	return nil
}

func (uc *updateJobUseCase) addJobStatusMetrics(prevJob *entities.Job, nextJobStatus entities.JobStatus) {
	uc.addMetrics(time.Since(prevJob.UpdatedAt), prevJob.Status, nextJobStatus, prevJob.ChainUUID)
}

func isValidJobStatus(nextStatus entities.JobStatus) bool {
	if nextStatus == entities.StatusResending {
		return false
	}
	if nextStatus == entities.StatusWarning {
		return false
	}
	if nextStatus == entities.StatusRecovering {
		return false
	}

	return true
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
