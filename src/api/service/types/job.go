package types

import (
	"time"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
)

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

type JobResponse struct {
	UUID          string                 `json:"uuid" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`                    // UUID of the job.
	ChainUUID     string                 `json:"chainUUID" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`               // UUID of the chain on which the job was created.
	ScheduleUUID  string                 `json:"scheduleUUID" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`            // UUID of the schedule on which the job was created.
	NextJobUUID   string                 `json:"nextJobUUID,omitempty" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`   // UUID of the next job.
	ParentJobUUID string                 `json:"parentJobUUID,omitempty" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"` // UUID of the parent job.
	TenantID      string                 `json:"tenantID" example:"foo"`                                                 // ID of the tenant executing the API.
	OwnerID       string                 `json:"ownerID,omitempty" example:"foo"`                                        // ID of the job owner.
	Transaction   ETHTransactionResponse `json:"transaction"`
	Logs          []*entities.Log        `json:"logs,omitempty"`   // List of logs.
	Labels        map[string]string      `json:"labels,omitempty"` // List of custom labels.
	Annotations   Annotations            `json:"annotations,omitempty"`
	Status        entities.JobStatus     `json:"status" example:"MINED"`                          // Status of the job.
	Type          entities.JobType       `json:"type" example:"eth://ethereum/transaction"`       // Type of job.
	CreatedAt     time.Time              `json:"createdAt" example:"2020-07-09T12:35:42.115395Z"` // Date and time at which the job was created.
	UpdatedAt     time.Time              `json:"updatedAt" example:"2020-07-09T12:35:42.115395Z"` // Date and time at which the job details were updated.
}

type Annotations struct {
	OneTimeKey     bool           `json:"oneTimeKey,omitempty" example:"true"`
	HasBeenRetried bool           `json:"hasBeenRetried,omitempty" example:"false"`
	GasPricePolicy GasPriceParams `json:"gasPricePolicy,omitempty"`
}

func (g *Annotations) Validate() error {
	if err := utils.GetValidator().Struct(g); err != nil {
		return err
	}

	if err := g.GasPricePolicy.RetryPolicy.Validate(); err != nil {
		return err
	}

	return nil
}
