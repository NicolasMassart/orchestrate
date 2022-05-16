package types

import (
	"github.com/consensys/orchestrate/src/entities"
)

type ContractEventMessageRequest struct {
	EventStream  *entities.EventStream  `json:"eventStream" validate:"required"`
	Notification *entities.Notification `json:"notification" validate:"required"`
}
