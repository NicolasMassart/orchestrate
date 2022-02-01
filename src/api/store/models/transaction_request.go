package models

import (
	"time"
)

type TransactionRequest struct {
	tableName struct{} `pg:"transaction_requests"` // nolint:unused,structcheck // reason

	ID             int
	IdempotencyKey string
	ChainName      string
	ScheduleID     *int
	Schedule       *Schedule
	RequestHash    string
	Params         interface{} `pg:",json"` // This will be automatically transformed in JSON by go-pg (and vice-versa)
	CreatedAt      time.Time   `pg:"default:now()"`
}
