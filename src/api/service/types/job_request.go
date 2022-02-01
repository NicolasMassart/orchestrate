package types

import "github.com/consensys/orchestrate/src/entities"

type CreateJobRequest struct {
	ScheduleUUID  string                  `json:"scheduleUUID" validate:"required,uuid4" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`           // UUID of the schedule on which to create the job.
	ChainUUID     string                  `json:"chainUUID" validate:"required,uuid4" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`              // UUID of the chain on which to create the job.
	NextJobUUID   string                  `json:"nextJobUUID,omitempty" validate:"omitempty,uuid4" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"` // UUID of the next job.
	Type          entities.JobType        `json:"type" validate:"required,isJobType" example:"eth://ethereum/transaction"`                         // Type of job.
	Labels        map[string]string       `json:"labels,omitempty"`                                                                                // List of custom labels.
	Annotations   Annotations             `json:"annotations,omitempty"`
	Transaction   entities.ETHTransaction `json:"transaction" validate:"required"`
	ParentJobUUID string                  `json:"parentJobUUID" validate:"omitempty,uuid4" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"` // UUID of the parent job.
}

type UpdateJobRequest struct {
	Labels      map[string]string        `json:"labels,omitempty"` // List of custom labels.
	Annotations *Annotations             `json:"annotations,omitempty"`
	Transaction *entities.ETHTransaction `json:"transaction,omitempty"`
	Status      entities.JobStatus       `json:"status,omitempty" validate:"isJobStatus" example:"MINED"` // Status of the job.
	Message     string                   `json:"message,omitempty" example:"Update message"`              // Update message.
}
