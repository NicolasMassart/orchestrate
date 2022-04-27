package jobs

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const failedSessionJobComponent = "tx-listener.use-case.tx-sentry.failed-job"

type failedJobUseCase struct {
	client orchestrateclient.OrchestrateClient
	logger *log.Logger
}

func FailedJobUseCase(client orchestrateclient.OrchestrateClient, logger *log.Logger) usecases.FailedJob {
	return &failedJobUseCase{
		client: client,
		logger: logger.SetComponent(failedSessionJobComponent),
	}
}

// Execute starts a job session
func (uc *failedJobUseCase) Execute(ctx context.Context, job *entities.Job, errMsg string) error {
	logger := uc.logger.WithField("job", job.UUID).WithField("reason", errMsg)
	logger.Debug("failed job")

	// Otherwise we failed on last job
	_, err := uc.client.UpdateJob(ctx, job.UUID, &types.UpdateJobRequest{
		Status:  entities.StatusFailed,
		Message: errMsg,
	})
	if err != nil {
		logger.WithError(err).Error("failed to update job")
		return errors.FromError(err).ExtendComponent(failedSessionJobComponent)
	}

	logger.Info("job has been updated to FAILED successfully")
	return nil
}
