package subscriptions

import (
	"context"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const notifySubscriptionEventsUseCaseComponent = "tx-listener.use-case.notify-subscription-events"

type notifySubscriptionEventsUC struct {
	subscriptionState store.Subscriptions
	messenger         sdk.MessengerAPI
	logger            *log.Logger
}

func NotifySubscriptionEventsUseCase(messenger sdk.MessengerAPI,
	subscriptionState store.Subscriptions,
	logger *log.Logger,
) usecases.NotifySubscriptionEvents {
	return &notifySubscriptionEventsUC{
		messenger:         messenger,
		subscriptionState: subscriptionState,
		logger:            logger.SetComponent(notifySubscriptionEventsUseCaseComponent),
	}
}

func (uc *notifySubscriptionEventsUC) Execute(ctx context.Context, chainUUID string, address ethcommon.Address, events []ethtypes.Log) error {
	logger := uc.logger.WithField("chain", chainUUID)

	logger.WithField("events_length", len(events)).Debug("handling subscriptions events")

	err := uc.messenger.ContractEventLogsMessage(ctx, &types.EventLogsMessageRequest{
		Address:   address,
		ChainUUID: chainUUID,
		EventLogs: events,
	}, multitenancy.NewInternalAdminUser())
	if err != nil {
		logger.WithError(err).Error("failed to persist subscription")
		return err
	}

	logger.Info("subscription events were sent successfully")
	return nil
}
