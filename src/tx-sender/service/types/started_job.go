package types

import (
	"github.com/consensys/orchestrate/src/entities"
)

type StartedJobReq struct {
	Job *entities.Job `json:"job"` // @TODO Add chain
}
