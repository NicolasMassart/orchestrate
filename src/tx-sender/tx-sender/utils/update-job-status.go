package utils

import (
	"context"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
)

func UpdateJobStatus(ctx context.Context, messenger sdk.MessengerAPI, job *entities.Job, status entities.JobStatus,
	msg string, transaction *entities.ETHTransaction) error {
	logger := log.FromContext(ctx).WithField("job", job.UUID).WithField("status", status)

	txUpdateReq := &api.JobUpdateMessageRequest{
		JobUUID: job.UUID,
		Status:  status,
		Message: msg,
	}

	if transaction != nil {
		txUpdateReq.Transaction = transaction
	}

	err := messenger.JobUpdateMessage(ctx, txUpdateReq, multitenancy.NewInternalAdminUser())
	if err != nil {
		logger.WithError(err).Error("failed to update job status")
		return err
	}

	job.Status = status
	logger.Info("job status update was sent successfully")
	return nil
}
