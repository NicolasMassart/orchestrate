package streams

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
)

const searchEventStreamsComponent = "use-cases.search-event_streams"

type searchUseCase struct {
	db     store.EventStreamAgent
	logger *log.Logger
}

func NewSearchUseCase(db store.EventStreamAgent) usecases.SearchEventStreamsUseCase {
	return &searchUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(searchEventStreamsComponent),
	}
}

func (uc *searchUseCase) Execute(ctx context.Context, filters *entities.EventStreamFilters, userInfo *multitenancy.UserInfo) ([]*entities.EventStream, error) {
	es, err := uc.db.Search(ctx, filters, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(searchEventStreamsComponent)
	}

	uc.logger.WithContext(ctx).Debug("event streams found successfully")
	return es, nil
}
