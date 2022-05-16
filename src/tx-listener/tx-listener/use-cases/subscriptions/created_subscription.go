package subscriptions

import (
	"context"
	"math/big"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const createdSubscriptionUseCaseComponent = "tx-listener.use-case.created-subscription"

type createdSubscriptionUC struct {
	ethClient         ethclient.Client
	proxyClient       sdk.ChainProxyClient
	subscriptionState store.Subscriptions
	messenger         sdk.MessengerAPI
	notifyEvents      usecases.NotifySubscriptionEvents
	logger            *log.Logger
}

func CreatedSubscriptionUseCase(messenger sdk.MessengerAPI,
	proxyClient sdk.ChainProxyClient,
	ethClient ethclient.Client,
	notifyEvents usecases.NotifySubscriptionEvents,
	subscriptionState store.Subscriptions,
	logger *log.Logger,
) usecases.CreatedSubscription {
	return &createdSubscriptionUC{
		messenger:         messenger,
		proxyClient:       proxyClient,
		ethClient:         ethClient,
		notifyEvents:      notifyEvents,
		subscriptionState: subscriptionState,
		logger:            logger.SetComponent(createdSubscriptionUseCaseComponent),
	}
}

func (uc *createdSubscriptionUC) Execute(ctx context.Context, sub *entities.Subscription) error {
	logger := uc.logger.WithField("subscription", sub.UUID).
		WithField("chain", sub.ChainUUID).
		WithField("address", sub.ContractAddress.String())

	logger.Debug("handling new subscriptions")

	if sub.FromBlock != nil {
		proxyURL := uc.proxyClient.ChainProxyURL(sub.ChainUUID)
		eventLogs, err := uc.ethClient.FilterLogs(ctx, proxyURL, []ethcommon.Address{sub.ContractAddress}, new(big.Int).SetUint64(*sub.FromBlock), nil)
		if err != nil {
			logger.WithError(err).Error("failed to query filtered logs")
			return err
		}
		err = uc.notifyEvents.Execute(ctx, sub.ChainUUID, sub.ContractAddress, eventLogs)
		if err != nil {
			return err
		}
	}

	err := uc.subscriptionState.Add(ctx, sub)
	if err != nil {
		logger.WithError(err).Error("failed to persist subscription")
		return err
	}

	return nil
}
