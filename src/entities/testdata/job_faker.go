package testdata

import (
	"math/big"
	"time"

	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/utils"

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
		ChainID:  big.NewInt(888),
		Priority: utils.PriorityMedium,
		StoreID:  "qkm-store-ID",
	}
}

type JobMatcher struct {
	Job *entities.Job
}

func NewJobMatcher(job *entities.Job) *JobMatcher {
	return &JobMatcher{
		Job: job,
	}
}

func (j *JobMatcher) Matches(x interface{}) bool {
	xJob, ok := x.(*entities.Job)
	if !ok {
		return false
	}
	if xJob.UUID != j.Job.UUID {
		return false
	}
	if xJob.ScheduleUUID != j.Job.ScheduleUUID {
		return false
	}
	if xJob.Status != j.Job.Status {
		return false
	}
	return true
}

func (j *JobMatcher) String() string {
	return j.Job.UUID
}

