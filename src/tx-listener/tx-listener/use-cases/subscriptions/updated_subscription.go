package subscriptions

import (
	"context"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const updatedSubscriptionUseCaseComponent = "tx-listener.use-case.updated-subscription"

type updatedSubscriptionUC struct {
	ethClient         ethclient.Client
	apiClient         sdk.OrchestrateClient
	subscriptionState store.Subscriptions
	messenger         sdk.MessengerAPI
	notifyEvents      usecases.NotifySubscriptionEvents
	logger            *log.Logger
}

func UpdatedSubscriptionUseCase(messenger sdk.MessengerAPI,
	apiClient sdk.OrchestrateClient,
	ethClient ethclient.Client,
	notifyEvents usecases.NotifySubscriptionEvents,
	subscriptionState store.Subscriptions,
	logger *log.Logger,
) usecases.UpdatedSubscription {
	return &updatedSubscriptionUC{
		messenger:         messenger,
		apiClient:         apiClient,
		ethClient:         ethClient,
		notifyEvents:      notifyEvents,
		subscriptionState: subscriptionState,
		logger:            logger.SetComponent(updatedSubscriptionUseCaseComponent),
	}
}

func (uc *updatedSubscriptionUC) Execute(ctx context.Context, sub *entities.Subscription) error {
	logger := uc.logger.WithField("subscription", sub.UUID).
		WithField("chain", sub.ChainUUID).
		WithField("address", sub.ContractAddress.String())

	logger.Debug("handling update of subscription")

	// @TODO Decide whether or not `fromBlock` can be updated
	// if sub.FromBlock != nil {
	// 	proxyURL := uc.apiClient.ChainProxyURL(sub.ChainUUID)
	// 	subQuery := ethereum.FilterQuery{
	// 		Addresses: []ethcommon.Address{sub.ContractAddress},
	// 		FromBlock: new(big.Int).SetUint64(*sub.FromBlock),
	// 	}
	// 	eventLogs, err := uc.ethClient.FilterLogs(ctx, proxyURL, subQuery)
	// 	if err != nil {
	// 		logger.WithError(err).Error("failed to query filtered logs")
	// 		return err
	// 	}
	// 	err = uc.notifyEvents.Execute(ctx, sub, eventLogs)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	err := uc.subscriptionState.Update(ctx, sub)
	if err != nil {
		logger.WithError(err).Error("failed to update subscription")
		return err
	}

	return nil
}
