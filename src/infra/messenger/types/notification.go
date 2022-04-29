package types

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
)

type NotificationResponse struct {
	SourceUUID string
	UUID       string
	Type       string
	Status     string
	APIVersion string
	Data       interface{}
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Error      string
}

type TxResponse struct { // TODO(dario): format TxResponse

}

func NewNotificationResponse(notif *entities.Notification) NotificationResponse {
	resp := NotificationResponse{
		SourceUUID: notif.SourceUUID,
		UUID:       notif.UUID,
		Type:       notif.Type.String(),
		Status:     notif.Status.String(),
		APIVersion: notif.APIVersion,
		CreatedAt:  notif.CreatedAt,
		UpdatedAt:  notif.UpdatedAt,
		Error:      notif.Error,
	}

	if notif.Type == entities.NotificationTypeTxMined {
		resp.Data = notif.Job // TODO(dario): Use TxResponse when formatted
	}

	return resp
}
