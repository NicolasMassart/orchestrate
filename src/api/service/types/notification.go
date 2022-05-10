package types

import (
	"github.com/consensys/orchestrate/src/entities"
)

type UpdateNotificationRequestMessage struct {
	UUID   string                      `json:"uuid,omitempty" validate:"required" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	Status entities.NotificationStatus `json:"status,omitempty" validate:"required,isNotificationStatus" example:"SENT"`
}
