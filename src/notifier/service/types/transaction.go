package types

import (
	"github.com/consensys/orchestrate/src/entities"
)

type TransactionMessageRequest struct {
	EventStream  *entities.EventStream  `json:"eventStream" validate:"required"`
	Notification *entities.Notification `json:"notification" validate:"required"`
}
