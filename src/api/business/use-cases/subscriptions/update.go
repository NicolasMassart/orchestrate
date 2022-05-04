package subscriptions

import (
	"context"
	"fmt"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
)

const updateSubscriptionComponent = "use-cases.update-subscription"

type updateUseCase struct {
	db                  store.SubscriptionAgent
	searchEventStreamUC usecases.SearchEventStreamsUseCase
	txListenerMessenger sdk.MessengerTxListener
	logger              *log.Logger
}

func NewUpdateUseCase(db store.SubscriptionAgent,
	searchEventStreamUC usecases.SearchEventStreamsUseCase,
	txListenerMessenger sdk.MessengerTxListener,
) usecases.UpdateSubscriptionUseCase {
	return &updateUseCase{
		db:                  db,
		searchEventStreamUC: searchEventStreamUC,
		txListenerMessenger: txListenerMessenger,
		logger:              log.NewLogger().SetComponent(updateSubscriptionComponent),
	}
}

func (uc *updateUseCase) Execute(ctx context.Context, subscription *entities.Subscription, eventStreamName string, userInfo *multitenancy.UserInfo) (*entities.Subscription, error) {
	ctx = log.WithFields(ctx, log.Field("subscription", subscription.UUID))
	logger := uc.logger.WithContext(ctx)

	logger.Debug("updating event stream")

	eventStream, err := uc.searchEventStreamUC.Execute(ctx, &entities.EventStreamFilters{Names: []string{eventStreamName}}, userInfo)
	if err != nil {
		return nil, err
	}
	if len(eventStream) == 0 {
		errMessage := fmt.Sprintf("event stream '%s' does not exist", eventStreamName)
		uc.logger.WithContext(ctx).Error(errMessage)
		return nil, errors.InvalidParameterError(errMessage).ExtendComponent(createSubscriptionComponent)
	}
	subscription.EventStreamUUID = eventStream[0].UUID

	_, err = uc.db.Update(ctx, subscription, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateSubscriptionComponent)
	}

	err = uc.txListenerMessenger.UpdateSubscriptionMessage(ctx, subscription, userInfo)
	if err != nil {
		errMsg := "failed to send update subscription message"
		uc.logger.WithError(err).Error(errMsg)
		return nil, errors.DependencyFailureError(errMsg).ExtendComponent(createSubscriptionComponent)
	}

	e, err := uc.db.FindOneByUUID(ctx, subscription.UUID, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(updateSubscriptionComponent)
	}

	logger.Info("event stream updated successfully")
	return e, nil
}
