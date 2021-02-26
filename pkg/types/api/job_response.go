package api

import (
	"time"

	"github.com/ConsenSys/orchestrate/pkg/types/entities"
)

type JobResponse struct {
	UUID          string                  `json:"uuid" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	ChainUUID     string                  `json:"chainUUID" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	ScheduleUUID  string                  `json:"scheduleUUID" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	NextJobUUID   string                  `json:"nextJobUUID,omitempty" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	ParentJobUUID string                  `json:"parentJobUUID,omitempty" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	TenantID      string                  `json:"tenantID,omitempty" example:"foo"`
	Transaction   entities.ETHTransaction `json:"transaction"`
	Logs          []*entities.Log         `json:"logs,omitempty"`
	Labels        map[string]string       `json:"labels,omitempty"`
	Annotations   Annotations             `json:"annotations,omitempty"`
	Status        entities.JobStatus      `json:"status" example:"MINED"`
	Type          entities.JobType        `json:"type" example:"eth://ethereum/transaction"`
	CreatedAt     time.Time               `json:"createdAt" example:"2020-07-09T12:35:42.115395Z"`
	UpdatedAt     time.Time               `json:"updatedAt" example:"2020-07-09T12:35:42.115395Z"`
}
