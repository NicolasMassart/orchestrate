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

const createEventStreamComponent = "use-cases.create-event_stream"

type createUseCase struct {
	db     store.EventStreamAgent
	logger *log.Logger
}

func NewCreateUseCase(db store.EventStreamAgent) usecases.CreateEventStreamUseCase {
	return &createUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(createEventStreamComponent),
	}
}

func (uc *createUseCase) Execute(ctx context.Context, eventStream *entities.EventStream, userInfo *multitenancy.UserInfo) (*entities.EventStream, error) {
	ctx = log.WithFields(ctx, log.Field("name", eventStream.Name))
	logger := uc.logger.WithContext(ctx)

	logger.Debug("creating new event stream")

	eventStreams, err := uc.db.Search(
		ctx,
		&entities.EventStreamFilters{Names: []string{eventStream.Name}, TenantID: userInfo.TenantID},
		userInfo.AllowedTenants,
		userInfo.Username,
	)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(createEventStreamComponent)
	}

	if len(eventStreams) > 0 {
		errMsg := "name already exists"
		logger.Error(errMsg)
		return nil, errors.AlreadyExistsError(errMsg).ExtendComponent(createEventStreamComponent)
	}

	e, err := uc.db.Insert(ctx, eventStream)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(createEventStreamComponent)
	}

	logger.WithField("name", e.Name).Info("event stream created successfully")
	return e, nil
}
