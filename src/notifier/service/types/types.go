package types

import "github.com/consensys/orchestrate/src/entities"

type NotificationMessageType string

const (
	TransactionNotificationType   NotificationMessageType = "transaction"
	ContractEventNotificationType NotificationMessageType = "contract_event"
)

type NotificationMessage struct {
	Type             NotificationMessageType `json:"type" validate:"required,isNotificationMessageType" example:"transaction"`
	EventStream      *entities.EventStream   `json:"eventStream" validate:"required"`
	SubscriptionUUID string                  `json:"subscriptionUUID,omitempty" validate:"omitempty" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	Job              *entities.Job           `json:"job,omitempty" validate:"omitempty"`
	Error            string                  `json:"error,omitempty" validate:"omitempty" example:"insufficient gas"`
}
