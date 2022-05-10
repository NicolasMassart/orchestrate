package models

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
)

type Notification struct {
	tableName struct{} `pg:"notifications"` // nolint:unused,structcheck // reason

	ID         int `pg:"alias:id"`
	UUID       string
	SourceUUID string
	SourceType string
	Status     string
	Type       string
	APIVersion string
	Error      string
	CreatedAt  time.Time `pg:"default:now()"`
	UpdatedAt  time.Time `pg:"default:now()"`
}

func NewNotification(notif *entities.Notification) *Notification {
	return &Notification{
		SourceUUID: notif.SourceUUID,
		SourceType: notif.SourceType.String(),
		Status:     notif.Status.String(),
		Type:       notif.Type.String(),
		APIVersion: notif.APIVersion,
		Error:      notif.Error,
	}
}

func (n *Notification) ToEntity() *entities.Notification {
	return &entities.Notification{
		UUID:       n.UUID,
		SourceUUID: n.SourceUUID,
		SourceType: entities.NotificationSourceType(n.SourceType),
		Status:     entities.NotificationStatus(n.Status),
		Type:       entities.NotificationType(n.Type),
		APIVersion: n.APIVersion,
		CreatedAt:  n.CreatedAt,
		UpdatedAt:  n.UpdatedAt,
		Error:      n.Error,
	}
}
