package schedules

import (
	"context"

	"github.com/ConsenSys/orchestrate/pkg/types/entities"
	usecases "github.com/ConsenSys/orchestrate/services/api/business/use-cases"

	"github.com/ConsenSys/orchestrate/pkg/errors"
	"github.com/ConsenSys/orchestrate/pkg/toolkit/app/log"
	"github.com/ConsenSys/orchestrate/services/api/business/parsers"
	"github.com/ConsenSys/orchestrate/services/api/store"
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
func (uc *searchSchedulesUseCase) Execute(ctx context.Context, tenants []string) ([]*entities.Schedule, error) {
	scheduleModels, err := uc.db.Schedule().FindAll(ctx, tenants)
	if err != nil {
		return nil, err
	}

	for idx, scheduleModel := range scheduleModels {
		for jdx, job := range scheduleModel.Jobs {
			scheduleModels[idx].Jobs[jdx], err = uc.db.Job().FindOneByUUID(ctx, job.UUID, tenants, false)
			if err != nil {
				return nil, errors.FromError(err).ExtendComponent(searchSchedulesComponent)
			}
		}
	}

	var resp []*entities.Schedule
	for _, s := range scheduleModels {
		resp = append(resp, parsers.NewScheduleEntityFromModels(s))
	}

	uc.logger.Info("schedules found successfully")
	return resp, nil
}
