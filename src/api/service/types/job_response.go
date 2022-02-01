package types

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
)

type JobResponse struct {
	UUID          string                  `json:"uuid" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`                    // UUID of the job.
	ChainUUID     string                  `json:"chainUUID" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`               // UUID of the chain on which the job was created.
	ScheduleUUID  string                  `json:"scheduleUUID" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`            // UUID of the schedule on which the job was created.
	NextJobUUID   string                  `json:"nextJobUUID,omitempty" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`   // UUID of the next job.
	ParentJobUUID string                  `json:"parentJobUUID,omitempty" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"` // UUID of the parent job.
	TenantID      string                  `json:"tenantID" example:"foo"`                                                 // ID of the tenant executing the API.
	OwnerID       string                  `json:"ownerID,omitempty" example:"foo"`                                        // ID of the job owner.
	Transaction   entities.ETHTransaction `json:"transaction"`
	Logs          []*entities.Log         `json:"logs,omitempty"`   // List of logs.
	Labels        map[string]string       `json:"labels,omitempty"` // List of custom labels.
	Annotations   Annotations             `json:"annotations,omitempty"`
	Status        entities.JobStatus      `json:"status" example:"MINED"`                          // Status of the job.
	Type          entities.JobType        `json:"type" example:"eth://ethereum/transaction"`       // Type of job.
	CreatedAt     time.Time               `json:"createdAt" example:"2020-07-09T12:35:42.115395Z"` // Date and time at which the job was created.
	UpdatedAt     time.Time               `json:"updatedAt" example:"2020-07-09T12:35:42.115395Z"` // Date and time at which the job details were updated.
}
