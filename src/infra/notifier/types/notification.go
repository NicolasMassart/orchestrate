package types

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
)

type NotificationType string

var (
	TransactionMinedMessage  NotificationType = "transaction.mined"
	TransactionFailedMessage NotificationType = "transaction.failed"
)

var jobStatusToNotificationType = map[entities.JobStatus]NotificationType{
	entities.StatusMined:  TransactionMinedMessage,
	entities.StatusFailed: TransactionFailedMessage,
}

func (n *NotificationType) String() string {
	return string(*n)
}

type Notification struct {
	UUID       string           `json:"uuid"`
	Type       NotificationType `json:"type"`
	APIVersion string           `json:"apiVersion"`
	Data       *TxResponse      `json:"data,omitempty"`
	CreatedAt  time.Time        `json:"createdAt"`
}

func NewTxNotification(job *entities.Job, errStr string) *Notification {
	return &Notification{
		UUID:       job.ScheduleUUID,
		Type:       jobStatusToNotificationType[job.Status],
		APIVersion: "v1",
		Data:       NewTxResponse(job, errStr),
		CreatedAt:  time.Now().UTC(),
	}
}
