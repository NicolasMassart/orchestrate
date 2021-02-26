package testutils

import (
	"time"

	"github.com/ConsenSys/orchestrate/pkg/types/entities"

	"github.com/ConsenSys/orchestrate/pkg/utils"

	"github.com/gofrs/uuid"
)

func FakeJob() *entities.Job {
	return &entities.Job{
		UUID:         uuid.Must(uuid.NewV4()).String(),
		ScheduleUUID: uuid.Must(uuid.NewV4()).String(),
		ChainUUID:    uuid.Must(uuid.NewV4()).String(),
		TenantID:     utils.RandString(6),
		Type:         entities.EthereumTransaction,
		InternalData: FakeInternalData(),
		Labels:       make(map[string]string),
		Logs:         []*entities.Log{FakeLog()},
		CreatedAt:    time.Now(),
		Status:       entities.StatusCreated,
		Transaction:  FakeETHTransaction(),
	}
}

func FakeInternalData() *entities.InternalData {
	return &entities.InternalData{
		ChainID:  "888",
		Priority: utils.PriorityMedium,
	}
}
