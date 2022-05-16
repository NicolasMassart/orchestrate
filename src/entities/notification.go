package entities

import (
	"time"

	"github.com/consensys/orchestrate/pkg/types/ethereum"
)

type NotificationType string
type NotificationStatus string
type NotificationSourceType string

const (
	NotificationTypeTxMined  NotificationType = "transaction.mined"
	NotificationTypeTxFailed NotificationType = "transaction.failed"
)
const (
	NotificationStatusPending NotificationStatus = "PENDING"
	NotificationStatusSent    NotificationStatus = "SENT"
)

const (
	NotificationSourceTypeJob           NotificationSourceType = "job"
	NotificationSourceTypeContractEvent NotificationSourceType = "contract_event"
)

func (n *NotificationType) String() string {
	return string(*n)
}

func (n *NotificationStatus) String() string {
	return string(*n)
}

func (n *NotificationSourceType) String() string {
	return string(*n)
}

// @TODO Refactor to decouple message types
type Notification struct {
	SourceUUID string
	SourceType NotificationSourceType
	Status     NotificationStatus
	UUID       string
	Type       NotificationType
	APIVersion string
	Job        *Job
	EventLogs  []*ethereum.Log
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Error      string
}
