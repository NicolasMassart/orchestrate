package postgres

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/infra/postgres"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/gofrs/uuid"
)

type PGLog struct {
	client postgres.Client
	logger *log.Logger
}

var _ store.LogAgent = &PGLog{}

func NewPGLog(client postgres.Client) *PGLog {
	return &PGLog{
		client: client,
		logger: log.NewLogger().SetComponent("data-agents.log"),
	}
}

func (agent *PGLog) Insert(ctx context.Context, logItem *entities.Log, jobUUID string) (*entities.Log, error) {
	model := models.NewLog(logItem)
	model.CreatedAt = time.Now().UTC()
	model.UUID = uuid.Must(uuid.NewV4()).String()

	job := &models.Job{}
	err := agent.client.ModelContext(ctx, job).Column("id").Where("uuid = ?", jobUUID).Select()
	if err != nil {
		errMsg := "failed to find corresponding job for the given log entry"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}
	model.JobID = &job.ID

	err = agent.client.ModelContext(ctx, model).Insert()
	if err != nil {
		errMessage := "failed to insert job log"
		agent.logger.WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return model.ToEntity(), nil
}
