package builder

import (
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/subscriptions"
)

type subscriptionsUCs struct {
	created      usecases.CreatedSubscription
	updated      usecases.UpdatedSubscription
	deleted      usecases.DeletedSubscription
	notifyEvents usecases.NotifySubscriptionEvents
}

func (b *subscriptionsUCs) CreatedSubscriptionUseCase() usecases.CreatedSubscription {
	return b.created
}

func (b *subscriptionsUCs) UpdatedSubscriptionUseCase() usecases.UpdatedSubscription {
	return b.updated
}

func (b *subscriptionsUCs) DeletedSubscriptionJobUseCase() usecases.DeletedSubscription {
	return b.deleted
}

func (b *subscriptionsUCs) NotifySubscriptionEventsUseCase() usecases.NotifySubscriptionEvents {
	return b.notifyEvents
}

func NewSubscriptionUseCase(messengerAPI sdk.MessengerAPI,
	apiClient sdk.OrchestrateClient,
	ethClient ethclient.MultiClient,
	subscriptionState store.Subscriptions,
	logger *log.Logger,
) usecases.SubscriptionUseCases {
	notifyEvents := subscriptions.NotifySubscriptionEventsUseCase(messengerAPI, subscriptionState, logger)
	createdUC := subscriptions.CreatedSubscriptionUseCase(messengerAPI, apiClient, ethClient, notifyEvents, subscriptionState, logger)
	updatedUC := subscriptions.UpdatedSubscriptionUseCase(messengerAPI, apiClient, ethClient, notifyEvents, subscriptionState, logger)
	deletedUC := subscriptions.DeletedSubscriptionUseCase(subscriptionState, logger)
	
	return &subscriptionsUCs{
		created:      createdUC,
		updated:      updatedUC,
		deleted:      deletedUC,
		notifyEvents: notifyEvents,
	}
}
