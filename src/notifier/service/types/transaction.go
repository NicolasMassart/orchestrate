package types

import (
	"github.com/consensys/orchestrate/src/entities"
)

type TransactionMessageRequest struct {
	EventStream *entities.EventStream `json:"eventStream" validate:"required"`
	Job         *entities.Job         `json:"job"`
	Error       string                `json:"error" example:"insufficient gas"`
}
