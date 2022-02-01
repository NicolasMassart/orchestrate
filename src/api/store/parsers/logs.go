package parsers

import (
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
)

func NewLogEntity(logModel *models.Log) *entities.Log {
	return &entities.Log{
		Status:    entities.JobStatus(logModel.Status),
		Message:   logModel.Message,
		CreatedAt: logModel.CreatedAt,
	}
}

func NewLogModel(log *entities.Log) *models.Log {
	return &models.Log{
		Status:    log.Status.String(),
		Message:   log.Message,
		CreatedAt: log.CreatedAt,
	}
}
