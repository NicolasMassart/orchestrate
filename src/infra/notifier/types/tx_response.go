package types

import (
	"github.com/consensys/orchestrate/src/entities"
)

type TxResponse struct {
	Job   *entities.Job `json:"job"` // TODO(dario): Create specific type for response instead of using the entity, similar to TxResponse in service layer
	Error string        `json:"error,omitempty"`
}

func NewTxResponse(job *entities.Job, errStr string) *TxResponse {
	return &TxResponse{
		Job:   job,
		Error: errStr,
	}
}
