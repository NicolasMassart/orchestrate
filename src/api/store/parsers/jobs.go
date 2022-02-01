package parsers

import (
	pkgjson "github.com/consensys/orchestrate/pkg/encoding/json"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
)

func NewJobModelFromEntities(job *entities.Job) *models.Job {
	jobModel := &models.Job{
		UUID:         job.UUID,
		ChainUUID:    job.ChainUUID,
		Type:         job.Type.String(),
		NextJobUUID:  job.NextJobUUID,
		Labels:       job.Labels,
		InternalData: job.InternalData,
		Status:       job.Status.String(),
		Schedule: &models.Schedule{
			UUID:     job.ScheduleUUID,
			TenantID: job.TenantID,
		},
		Logs:      []*models.Log{},
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
		jobModel.Transaction = NewTransactionModel(job.Transaction)
	}

	for _, log := range job.Logs {
		jobModel.Logs = append(jobModel.Logs, NewLogModel(log))
	}

	return jobModel
}

func NewJobEntity(jobModel *models.Job) *entities.Job {
	job := &entities.Job{
		UUID:        jobModel.UUID,
		ChainUUID:   jobModel.ChainUUID,
		NextJobUUID: jobModel.NextJobUUID,
		Type:        entities.JobType(jobModel.Type),
		Labels:      jobModel.Labels,
		Logs:        []*entities.Log{},
		CreatedAt:   jobModel.CreatedAt,
		UpdatedAt:   jobModel.UpdatedAt,
		Status:      entities.JobStatus(jobModel.Status),
	}

	if jobModel.InternalData != nil {
		job.InternalData = &entities.InternalData{}
		_ = pkgjson.UnmarshalInterface(jobModel.InternalData, job.InternalData)
	}

	if jobModel.Schedule != nil {
		job.ScheduleUUID = jobModel.Schedule.UUID
		job.TenantID = jobModel.Schedule.TenantID
		job.OwnerID = jobModel.Schedule.OwnerID
	}

	if jobModel.Transaction != nil {
		job.Transaction = NewTransactionEntity(jobModel.Transaction)
	}

	for _, logModel := range jobModel.Logs {
		job.Logs = append(job.Logs, NewLogEntity(logModel))
	}

	return job
}

func NewJobEntityArr(jobs []*models.Job) []*entities.Job {
	res := []*entities.Job{}
	for _, j := range jobs {
		res = append(res, NewJobEntity(j))
	}
	return res
}
