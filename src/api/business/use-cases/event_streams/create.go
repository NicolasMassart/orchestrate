package streams

import (
	"context"
	"fmt"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
)

const createEventStreamComponent = "use-cases.create-event_stream"

type createUseCase struct {
	db             store.EventStreamAgent
	searchChainsUC usecases.SearchChainsUseCase
	logger         *log.Logger
}

func NewCreateUseCase(db store.EventStreamAgent, searchChainsUC usecases.SearchChainsUseCase) usecases.CreateEventStreamUseCase {
	return &createUseCase{
		db:             db,
		searchChainsUC: searchChainsUC,
		logger:         log.NewLogger().SetComponent(createEventStreamComponent),
	}
}

func (uc *createUseCase) Execute(ctx context.Context, eventStream *entities.EventStream, chainName string, userInfo *multitenancy.UserInfo) (*entities.EventStream, error) {
	ctx = log.WithFields(ctx, log.Field("name", eventStream.Name))
	logger := uc.logger.WithContext(ctx)

	logger.Debug("creating new event stream")

	eventStreams, err := uc.db.Search(ctx,
		&entities.EventStreamFilters{Names: []string{eventStream.Name}, TenantID: userInfo.TenantID},
		userInfo.AllowedTenants,
		userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(createEventStreamComponent)
	}
	if len(eventStreams) > 0 {
		errMsg := "event stream with same name already exists"
		logger.Error(errMsg)
		return nil, errors.AlreadyExistsError(errMsg).ExtendComponent(createEventStreamComponent)
	}

	if chainName != "" {
		chains, err2 := uc.searchChainsUC.Execute(ctx, &entities.ChainFilters{Names: []string{chainName}}, userInfo)
		if err2 != nil {
			return nil, err2
		}

		if len(chains) == 0 {
			errMessage := fmt.Sprintf("chain '%s' does not exist", chainName)
			uc.logger.WithContext(ctx).Error(errMessage)
			return nil, errors.InvalidParameterError(errMessage).ExtendComponent(createEventStreamComponent)
		}

		eventStream.ChainUUID = chains[0].UUID
	}

	eventStreams, err = uc.db.Search(ctx,
		&entities.EventStreamFilters{ChainUUID: eventStream.ChainUUID, TenantID: userInfo.TenantID},
		userInfo.AllowedTenants,
		userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(createEventStreamComponent)
	}
	if len(eventStreams) > 0 {
		errMsg := "multiple event streams on notification events"
		logger.Error(errMsg)
		return nil, errors.AlreadyExistsError(errMsg).ExtendComponent(createEventStreamComponent)
	}

	eventStream.TenantID = userInfo.TenantID
	eventStream.OwnerID = userInfo.Username
	e, err := uc.db.Insert(ctx, eventStream)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(createEventStreamComponent)
	}

	logger.WithField("event_stream", e.UUID).Info("event stream created successfully")
	return e, nil
}
