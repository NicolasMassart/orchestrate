package entities

import (
	"time"
)

type NotificationType string
type NotificationStatus string

const (
	NotificationTypeTxMined  NotificationType = "transaction.mined"
	NotificationTypeTxFailed NotificationType = "transaction.failed"
)
const (
	NotificationStatusPending NotificationStatus = "PENDING"
	NotificationStatusSent    NotificationStatus = "SENT"
)

func (n *NotificationType) String() string {
	return string(*n)
}

func (n *NotificationStatus) String() string {
	return string(*n)
}

type Notification struct {
	SourceUUID string
	Status     NotificationStatus
	UUID       string
	Type       NotificationType
	APIVersion string
	Job        *Job
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Error      string
}
