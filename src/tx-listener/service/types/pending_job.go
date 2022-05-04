package types

import (
	"github.com/consensys/orchestrate/src/entities"
)

type PendingJobMessageRequest struct {
	Job *entities.Job `json:"job,omitempty"` // @TODO Add chain
}
