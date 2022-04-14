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

const updateEventStreamComponent = "use-cases.update-event_stream"

type updateUseCase struct {
	db             store.EventStreamAgent
	searchChainsUC usecases.SearchChainsUseCase
	logger         *log.Logger
}

func NewUpdateUseCase(db store.EventStreamAgent, searchChainsUC usecases.SearchChainsUseCase) usecases.UpdateEventStreamUseCase {
	return &updateUseCase{
		db:             db,
		searchChainsUC: searchChainsUC,
		logger:         log.NewLogger().SetComponent(updateEventStreamComponent),
	}
}

func (uc *updateUseCase) Execute(ctx context.Context, eventStream *entities.EventStream, chainName string, userInfo *multitenancy.UserInfo) (*entities.EventStream, error) {
	ctx = log.WithFields(ctx, log.Field("name", eventStream.Name), log.Field("event_stream_uuid", eventStream.UUID))
	logger := uc.logger.WithContext(ctx)

	logger.Debug("updating event stream")

	filter := &entities.EventStreamFilters{Names: []string{eventStream.Name}, TenantID: userInfo.TenantID}
	if chainName != "" {
		chains, err := uc.searchChainsUC.Execute(ctx, &entities.ChainFilters{Names: []string{chainName}}, userInfo)
		if err != nil {
			return nil, err
		}

		if len(chains) == 0 {
			errMessage := fmt.Sprintf("chain '%s' does not exist", chainName)
			uc.logger.WithContext(ctx).Error(errMessage)
			return nil, errors.InvalidParameterError(errMessage).ExtendComponent(updateEventStreamComponent)
		}

		eventStream.ChainUUID = chains[0].UUID
		filter.ChainUUID = chains[0].UUID
	}

	eventStreams, err := uc.db.Search(ctx, filter, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateEventStreamComponent)
	}

	if len(eventStreams) > 0 {
		errMsg := "event stream with same name already exists"
		logger.Error(errMsg)
		return nil, errors.AlreadyExistsError(errMsg).ExtendComponent(updateEventStreamComponent)
	}

	_, err = uc.db.Update(ctx, eventStream, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateEventStreamComponent)
	}

	e, err := uc.db.FindOneByUUID(ctx, eventStream.UUID, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateEventStreamComponent)
	}

	logger.Info("event stream updated successfully")
	return e, nil
}
