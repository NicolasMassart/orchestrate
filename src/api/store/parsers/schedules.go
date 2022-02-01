package parsers

import (
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
)

func NewScheduleEntity(scheduleModel *models.Schedule) *entities.Schedule {
	schedule := &entities.Schedule{
		UUID:      scheduleModel.UUID,
		TenantID:  scheduleModel.TenantID,
		OwnerID:   scheduleModel.OwnerID,
		CreatedAt: scheduleModel.CreatedAt,
	}

	for _, job := range scheduleModel.Jobs {
		schedule.Jobs = append(schedule.Jobs, NewJobEntity(job))
	}

	return schedule
}

func NewScheduleEntityArr(schedules []*models.Schedule) []*entities.Schedule {
	res := []*entities.Schedule{}
	for _, s := range schedules {
		res = append(res, NewScheduleEntity(s))
	}
	return res
}

func NewScheduleModel(schedule *entities.Schedule) *models.Schedule {
	scheduleModel := &models.Schedule{
		UUID:     schedule.UUID,
		TenantID: schedule.TenantID,
	}

	for _, job := range schedule.Jobs {
		jobModel := NewJobModelFromEntities(job)
		jobModel.ScheduleID = &scheduleModel.ID
		scheduleModel.Jobs = append(scheduleModel.Jobs, NewJobModelFromEntities(job))
	}

	return scheduleModel
}
