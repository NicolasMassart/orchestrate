package models

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
)

type Job struct {
	tableName struct{} `pg:"jobs"` // nolint:unused,structcheck // reason

	ID            int `pg:"alias:id"`
	UUID          string
	ChainUUID     string
	NextJobUUID   string    `pg:"alias:next_job_uuid"`
	ScheduleID    *int      `pg:"alias:schedule_id,notnull"`
	Schedule      *Schedule `pg:"rel:has-one"`
	Type          string
	TransactionID *int         `pg:"alias:transaction_id,notnull"`
	Transaction   *Transaction `pg:"rel:has-one"`
	Logs          []*Log       `pg:"rel:has-many"`
	Labels        map[string]string
	InternalData  *entities.InternalData
	IsParent      bool `pg:"alias:is_parent,default:false,use_zero"`
	Status        string
	CreatedAt     time.Time `pg:"default:now()"`
	UpdatedAt     time.Time `pg:"default:now()"`
}

func NewJob(job *entities.Job) *Job {
	jobModel := &Job{
		UUID:         job.UUID,
		ChainUUID:    job.ChainUUID,
		Type:         job.Type.String(),
		NextJobUUID:  job.NextJobUUID,
		Labels:       job.Labels,
		InternalData: job.InternalData,
		Status:       job.Status.String(),
		Schedule: &Schedule{
			UUID:     job.ScheduleUUID,
			TenantID: job.TenantID,
		},
		Logs:      []*Log{},
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
	}

	if job.InternalData != nil {
		jobModel.IsParent = job.InternalData.ParentJobUUID == ""
	}

	if job.Status == "" {
		job.Status = entities.StatusCreated
	}

	if job.Transaction != nil {
		jobModel.Transaction = NewTransaction(job.Transaction)
	}

	for _, log := range job.Logs {
		jobModel.Logs = append(jobModel.Logs, NewLog(log))
	}

	return jobModel
}

func NewJobs(jobs []*Job) []*entities.Job {
	res := []*entities.Job{}
	for _, j := range jobs {
		res = append(res, j.ToEntity())
	}

	return res
}

func (j *Job) ToEntity() *entities.Job {
	job := &entities.Job{
		UUID:         j.UUID,
		ChainUUID:    j.ChainUUID,
		NextJobUUID:  j.NextJobUUID,
		Type:         entities.JobType(j.Type),
		Labels:       j.Labels,
		Logs:         []*entities.Log{},
		InternalData: j.InternalData,
		CreatedAt:    j.CreatedAt,
		UpdatedAt:    j.UpdatedAt,
		Status:       entities.JobStatus(j.Status),
	}

	if j.Schedule != nil {
		job.ScheduleUUID = j.Schedule.UUID
		job.TenantID = j.Schedule.TenantID
		job.OwnerID = j.Schedule.OwnerID
	}

	if j.Transaction != nil {
		job.Transaction = j.Transaction.ToEntity()
	}

	for _, log := range j.Logs {
		job.Logs = append(job.Logs, log.ToEntity())
	}

	return job
}
