package streams

import (
	"context"
	"encoding/json"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const notifyContractEventsComponent = "use-cases.notify_contract_events"

type notifyContractEventsUseCase struct {
	db                store.DB
	searchContractUC  usecases.SearchContractUseCase
	decodeLogUC       usecases.DecodeEventLogUseCase
	notifierMessenger sdk.MessengerNotifier
	logger            *log.Logger
}

func NewNotifyContractEventsUseCase(
	db store.DB,
	searchContractUC usecases.SearchContractUseCase,
	decodeLogUC usecases.DecodeEventLogUseCase,
	notifierMessenger sdk.MessengerNotifier,
) usecases.NotifyContractEventsUseCase {
	return &notifyContractEventsUseCase{
		db:                db,
		searchContractUC:  searchContractUC,
		decodeLogUC:       decodeLogUC,
		notifierMessenger: notifierMessenger,
		logger:            log.NewLogger().SetComponent(notifyContractEventsComponent),
	}
}

// @TODO Make it based on chainID
func (uc *notifyContractEventsUseCase) Execute(ctx context.Context, chainUUID string, address ethcommon.Address,
	eventLogs []ethtypes.Log, userInfo *multitenancy.UserInfo) error {
	logger := uc.logger.WithField("chain", chainUUID).WithField("address", address.String())

	logger.Debug("building notification")
	subscriptions, err := uc.db.Subscription().Search(ctx,
		&entities.SubscriptionFilters{ChainUUID: chainUUID, Addresses: []ethcommon.Address{address}},
		userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return err
	}

	if len(subscriptions) == 0 {
		logger.Warn("not subscription were found")
		return nil
	}

	for _, sub := range subscriptions {
		eventStream, err := uc.db.EventStream().FindOneByUUID(ctx, sub.EventStreamUUID, userInfo.AllowedTenants, userInfo.Username)
		if err != nil {
			return errors.FromError(err).ExtendComponent(notifyTransactionComponent)
		}

		// @TODO Restore after contract registry refactor
		decodedEventLogs := []*ethereum.Log{}
		for eventLogIdx := range eventLogs {
			l := &ethereum.Log{}
			bEventLog, _ := json.Marshal(eventLogs[eventLogIdx])
			_ = json.Unmarshal(bEventLog, l)
			decodedLog, err2 := uc.decodeLogUC.Execute(ctx, sub.ChainUUID, l)
			if err2 != nil {
				return errors.FromError(err2).ExtendComponent(notifyContractEventsComponent)
			}
			if decodedLog != nil {
				decodedEventLogs = append(decodedEventLogs, decodedLog)
			}
		}

		notif, err := uc.db.Notification().Insert(ctx, &entities.Notification{
			SourceUUID: sub.UUID,
			SourceType: entities.NotificationSourceTypeContractEvent,
			Status:     entities.NotificationStatusPending,
			APIVersion: "v1",
		})
		if err != nil {
			return errors.FromError(err).ExtendComponent(notifyContractEventsComponent)
		}
		notif.EventLogs = decodedEventLogs

		if eventStream.Status == entities.EventStreamStatusLive {
			err = uc.notifierMessenger.ContractEventNotificationMessage(ctx, eventStream, notif, userInfo)
			if err != nil {
				errMsg := "failed to send contract event notification"
				logger.WithError(err).Error(errMsg)
				return errors.DependencyFailureError(errMsg).ExtendComponent(notifyContractEventsComponent)
			}
		}

		logger.WithField("subscription", sub.UUID).Debug("notification sent successfully")
	}

	return nil
}
