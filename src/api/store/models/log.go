package models

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
)

type Log struct {
	tableName struct{} `pg:"logs"` // nolint:unused,structcheck // reason

	ID        int `pg:"alias:id"`
	UUID      string
	JobID     *int `pg:"alias:job_id,notnull"`
	Job       *Job `pg:"rel:has-one"`
	Status    string
	Message   string
	CreatedAt time.Time `pg:"default:now()"`
}

func NewLog(log *entities.Log) *Log {
	return &Log{
		Status:    log.Status.String(),
		Message:   log.Message,
		CreatedAt: log.CreatedAt,
	}
}

func (log *Log) ToEntity() *entities.Log {
	return &entities.Log{
		Status:    entities.JobStatus(log.Status),
		Message:   log.Message,
		CreatedAt: log.CreatedAt,
	}
}
