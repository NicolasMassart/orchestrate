package utils

import (
	"context"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
)

func UpdateJobStatus(ctx context.Context, apiClient client.JobClient, job *entities.Job, status entities.JobStatus,
	msg string, transaction *entities.ETHTransaction) error {
	logger := log.FromContext(ctx).WithField("status", status)

	txUpdateReq := &api.UpdateJobRequest{
		Status:      status,
		Message:     msg,
		Transaction: transaction,
	}

	_, err := apiClient.UpdateJob(ctx, job.UUID, txUpdateReq)
	if err != nil {
		logger.WithError(err).Error("failed to update job status")
		return err
	}

	job.Status = status
	logger.Debug("job status was updated successfully")
	return nil
}
