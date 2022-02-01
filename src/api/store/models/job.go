package models

import (
	"time"
)

type Job struct {
	tableName struct{} `pg:"jobs"` // nolint:unused,structcheck // reason

	ID            int `pg:"alias:id"`
	UUID          string
	ChainUUID     string
	NextJobUUID   string `pg:"alias:next_job_uuid"`
	ScheduleID    *int   `pg:"alias:schedule_id,notnull"`
	Schedule      *Schedule
	Type          string
	TransactionID *int `pg:"alias:transaction_id,notnull"`
	Transaction   *Transaction
	Logs          []*Log
	Labels        map[string]string
	InternalData  interface{} `pg:",json"`
	IsParent      bool        `pg:"alias:is_parent,default:false,use_zero"`
	Status        string
	CreatedAt     time.Time `pg:"default:now()"`
	UpdatedAt     time.Time `pg:"default:now()"`
}
