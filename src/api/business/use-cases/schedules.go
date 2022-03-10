package usecases

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=schedules.go -destination=mocks/schedules.go -package=mocks

type ScheduleUseCases interface {
	CreateSchedule() CreateScheduleUseCase
	GetSchedule() GetScheduleUseCase
	SearchSchedules() SearchSchedulesUseCase
}

type CreateScheduleUseCase interface {
	Execute(ctx context.Context, schedule *entities.Schedule, userInfo *multitenancy.UserInfo) (*entities.Schedule, error)
}

type GetScheduleUseCase interface {
	Execute(ctx context.Context, scheduleUUID string, userInfo *multitenancy.UserInfo) (*entities.Schedule, error)
}

type SearchSchedulesUseCase interface {
	Execute(ctx context.Context, userInfo *multitenancy.UserInfo) ([]*entities.Schedule, error)
}
