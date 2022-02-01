package schedules

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
)

const searchSchedulesComponent = "use-cases.search-schedules"

// searchSchedulesUseCase is a use case to search schedules
type searchSchedulesUseCase struct {
	db     store.DB
	logger *log.Logger
}

// NewSearchSchedulesUseCase creates a new SearchSchedulesUseCase
func NewSearchSchedulesUseCase(db store.DB) usecases.SearchSchedulesUseCase {
	return &searchSchedulesUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(searchSchedulesComponent),
	}
}

// Execute search schedules
func (uc *searchSchedulesUseCase) Execute(ctx context.Context, userInfo *multitenancy.UserInfo) ([]*entities.Schedule, error) {
	schedules, err := uc.db.Schedule().FindAll(ctx, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, err
	}

	for idx, scheduleModel := range schedules {
		for jdx, job := range scheduleModel.Jobs {
			schedules[idx].Jobs[jdx], err = uc.db.Job().FindOneByUUID(ctx, job.UUID, userInfo.AllowedTenants, userInfo.Username, false)
			if err != nil {
				return nil, errors.FromError(err).ExtendComponent(searchSchedulesComponent)
			}
		}
	}

	uc.logger.Info("schedules found successfully")
	return schedules, nil
}
