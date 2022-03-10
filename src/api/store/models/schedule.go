package models

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
)

type Schedule struct {
	tableName struct{} `pg:"schedules"` // nolint:unused,structcheck // reason

	ID        int `pg:"alias:id"`
	UUID      string
	TenantID  string    `pg:"alias:tenant_id"`
	OwnerID   string    `pg:"alias:owner_id"`
	Jobs      []*Job    `pg:"rel:has-many"`
	CreatedAt time.Time `pg:"default:now()"`
}

func NewSchedule(schedule *entities.Schedule) *Schedule {
	scheduleModel := &Schedule{
		UUID:     schedule.UUID,
		TenantID: schedule.TenantID,
		OwnerID:  schedule.OwnerID,
	}

	for _, job := range schedule.Jobs {
		jobModel := NewJob(job)
		jobModel.ScheduleID = &scheduleModel.ID
		scheduleModel.Jobs = append(scheduleModel.Jobs, NewJob(job))
	}

	return scheduleModel
}

func NewSchedules(schedules []Schedule) []*entities.Schedule {
	res := []*entities.Schedule{}
	for _, s := range schedules {
		res = append(res, s.ToEntity())
	}

	return res
}

func (s *Schedule) ToEntity() *entities.Schedule {
	schedule := &entities.Schedule{
		UUID:      s.UUID,
		TenantID:  s.TenantID,
		OwnerID:   s.OwnerID,
		CreatedAt: s.CreatedAt,
	}

	for _, job := range s.Jobs {
		schedule.Jobs = append(schedule.Jobs, job.ToEntity())
	}

	return schedule
}
