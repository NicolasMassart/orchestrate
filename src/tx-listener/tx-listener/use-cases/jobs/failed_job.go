package jobs

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const failedSessionJobComponent = "tx-listener.use-case.failed-job"

type failedJobUseCase struct {
	completedJob    usecases.CompletedJob
	messenger       sdk.MessengerAPI
	pendingJobState store.PendingJob
	logger          *log.Logger
}

func FailedJobUseCase(messengerCli sdk.MessengerAPI, commitMessage usecases.CompletedJob,
	pendingJobState store.PendingJob, logger *log.Logger) usecases.FailedJob {
	return &failedJobUseCase{
		completedJob:    commitMessage,
		messenger:       messengerCli,
		pendingJobState: pendingJobState,
		logger:          logger.SetComponent(failedSessionJobComponent),
	}
}

// Execute starts a job session
func (uc *failedJobUseCase) Execute(ctx context.Context, job *entities.Job, errMsg string) error {
	logger := uc.logger.WithField("job", job.UUID).WithField("reason", errMsg)
	logger.Debug("failed job")

	// Otherwise we failed on last job
	err := uc.messenger.JobUpdateMessage(ctx, &types.JobUpdateMessageRequest{
		JobUUID: job.UUID,
		Status:  entities.StatusFailed,
		Message: errMsg,
	}, multitenancy.NewInternalAdminUser())
	if err != nil {
		logger.WithError(err).Error("failed to update job")
		return errors.FromError(err).ExtendComponent(failedSessionJobComponent)
	}

	err = uc.completedJob.Execute(ctx, job)
	if err != nil {
		return err
	}

	logger.Info("job has been updated to FAILED successfully")
	return nil
}
