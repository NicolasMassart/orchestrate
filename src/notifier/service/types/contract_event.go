package types

import (
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/src/entities"
)

type ContractEventMessageRequest struct {
	EventStream      *entities.EventStream `json:"eventStream" validate:"required"`
	SubscriptionUUID string                `json:"subscriptionUUID" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	EventLogs        []*ethereum.Log       `json:"event_logs" validate:"omitempty"`
}
