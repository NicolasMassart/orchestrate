package jobs

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
)

const searchJobsComponent = "use-cases.search-jobs"

// searchJobsUseCase is a use case to search jobs
type searchJobsUseCase struct {
	db     store.DB
	logger *log.Logger
}

// NewSearchJobsUseCase creates a new SearchJobsUseCase
func NewSearchJobsUseCase(db store.DB) usecases.SearchJobsUseCase {
	return &searchJobsUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(searchJobsComponent),
	}
}

// Execute search jobs
func (uc *searchJobsUseCase) Execute(ctx context.Context, filters *entities.JobFilters, userInfo *multitenancy.UserInfo) ([]*entities.Job, error) {
	jobs, err := uc.db.Job().Search(ctx, filters, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(searchJobsComponent)
	}

	uc.logger.Trace("jobs found successfully")
	return jobs, nil
}
