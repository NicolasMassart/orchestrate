package dataagents

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/parsers"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	pg "github.com/consensys/orchestrate/src/infra/database/postgres"
	"github.com/gofrs/uuid"
)

const logDAComponent = "data-agents.log"

// PGLog is a log data agent for PostgreSQL
type PGLog struct {
	db     pg.DB
	logger *log.Logger
}

// NewPGLog creates a new PGLog
func NewPGLog(db pg.DB) store.LogAgent {
	return &PGLog{db: db, logger: log.NewLogger().SetComponent(logDAComponent)}
}

// Insert Inserts a new log in DB
func (agent *PGLog) Insert(ctx context.Context, logItem *entities.Log, jobUUID string) error {
	model := parsers.NewLogModel(logItem)
	if model.UUID == "" {
		model.UUID = uuid.Must(uuid.NewV4()).String()
	}
	model.CreatedAt = time.Now().UTC()

	jobID, err := getJobIDByUUID(ctx, agent.db, agent.logger, jobUUID)
	if err != nil {
		return errors.FromError(err).ExtendComponent(jobDAComponent)
	}
	model.JobID = &jobID

	err = pg.Insert(ctx, agent.db, model)
	if err != nil {
		agent.logger.WithError(err).Error("failed to insert job log")
		return errors.FromError(err).ExtendComponent(logDAComponent)
	}

	utils.CopyPtr(parsers.NewLogEntity(model), logItem)
	return nil
}
