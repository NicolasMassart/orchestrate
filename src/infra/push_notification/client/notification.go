package client

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
	"github.com/gofrs/uuid"
)

var jobStatusToNotificationType = map[entities.JobStatus]string{
	entities.StatusMined:  "transaction.mined",
	entities.StatusFailed: "transaction.failed",
}

type Notification struct {
	UUID       string      `json:"uuid"`
	Type       string      `json:"type"`
	APIVersion string      `json:"apiVersion"`
	Data       interface{} `json:"data,omitempty"`
	CreatedAt  time.Time   `json:"createdAt"`
}

func NewTxNotification(job *entities.Job, errStr string) *Notification {
	return &Notification{
		UUID:       uuid.Must(uuid.NewV4()).String(),
		Type:       jobStatusToNotificationType[job.Status],
		APIVersion: "v1",
		Data:       NewTxResponse(job, errStr),
		CreatedAt:  time.Now().UTC(),
	}
}
