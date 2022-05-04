package subscriptions

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const notifySubscriptionEventComponent = "use-cases.notify_subscription_events"

type notifySubscriptionEventUseCase struct {
	db                   store.SubscriptionAgent
	searchSubscriptionUC usecases.SearchSubscriptionUseCase
	searchContractUC     usecases.SearchContractUseCase
	getEventStreamUC     usecases.GetEventStreamUseCase
	decodeLogUC          usecases.DecodeEventLogUseCase
	notifierMessenger    sdk.MessengerNotifier
	logger               *log.Logger
}

func NewNotifySubscriptionEventUseCase(
	db store.SubscriptionAgent,
	searchSubscriptionUC usecases.SearchSubscriptionUseCase,
	searchContractUC usecases.SearchContractUseCase,
	getEventStreamUC usecases.GetEventStreamUseCase,
	decodeLogUC usecases.DecodeEventLogUseCase,
	notifierMessenger sdk.MessengerNotifier,
) usecases.NotifySubscriptionEventUseCase {
	return &notifySubscriptionEventUseCase{
		db:                   db,
		searchSubscriptionUC: searchSubscriptionUC,
		searchContractUC:     searchContractUC,
		getEventStreamUC:     getEventStreamUC,
		decodeLogUC:          decodeLogUC,
		notifierMessenger:    notifierMessenger,
		logger:               log.NewLogger().SetComponent(notifySubscriptionEventComponent),
	}
}

// @TODO Make it based on chainID
func (uc *notifySubscriptionEventUseCase) Execute(ctx context.Context, chainUUID string, address ethcommon.Address,
	eventLogs []*ethereum.Log, userInfo *multitenancy.UserInfo) error {
	logger := uc.logger.WithField("chain", chainUUID).WithField("address", address.String())

	subscriptions, err := uc.searchSubscriptionUC.Execute(ctx,
		&entities.SubscriptionFilters{ChainUUID: chainUUID, Addresses: []ethcommon.Address{address}},
		userInfo)
	if err != nil {
		return err
	}

	for _, sub := range subscriptions {
		eventStream, err := uc.getEventStreamUC.Execute(ctx, sub.EventStreamUUID, userInfo)
		if err != nil {
			return err
		}
		eventLogs := []*ethereum.Log{}
		for _, l := range eventLogs {
			decodedLog, err2 := uc.decodeLogUC.Execute(ctx, sub.ChainUUID, l)
			if err2 != nil {
				return errors.FromError(err2).ExtendComponent(notifySubscriptionEventComponent)
			}
			if decodedLog != nil {
				eventLogs = append(eventLogs, decodedLog)
			}
		}

		err = uc.notifierMessenger.ContractEventNotificationMessage(ctx, eventStream, sub.UUID, eventLogs, userInfo)
		if err != nil {
			errMsg := "failed to send notification to notifier service"
			logger.WithError(err).Error(errMsg)
			return errors.DependencyFailureError(errMsg).ExtendComponent(notifySubscriptionEventComponent)
		}

		uc.logger.WithField("subscription", sub.UUID).Debug("notification sent successfully")
	}

	return nil
}
