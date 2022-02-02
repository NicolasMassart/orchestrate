package types

import (
	"time"
)

type CreateScheduleRequest struct{}

type ScheduleResponse struct {
	UUID      string         `json:"uuid" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"` // UUID of the schedule.
	TenantID  string         `json:"tenantID" example:"tenant_id"`                        // ID of the tenant executing the API.
	OwnerID   string         `json:"ownerID,omitempty" example:"foo"`                     // ID of the schedule owner.
	Jobs      []*JobResponse `json:"jobs"`                                                // List of jobs on the schedule.
	CreatedAt time.Time      `json:"createdAt" example:"2020-07-09T12:35:42.115395Z"`     // Date and time at which the schedule was created.
}
