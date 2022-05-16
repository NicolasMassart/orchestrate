package chains

import (
	"context"
	"math/big"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const chainBlockEventsUseCaseComponent = "tx-listener.use-case.chain-block-events"

type chainBlockEventsUC struct {
	ethClient         ethclient.Client
	proxyClient       sdk.ChainProxyClient
	subscriptionState store.Subscriptions
	notifyEvents      usecases.NotifySubscriptionEvents
	logger            *log.Logger
}

func NewChainBlockEventsUseCase(proxyClient sdk.ChainProxyClient,
	ethClient ethclient.Client,
	notifyEvents usecases.NotifySubscriptionEvents,
	subscriptionState store.Subscriptions,
	logger *log.Logger,
) usecases.ChainBlockEvents {
	return &chainBlockEventsUC{
		proxyClient:       proxyClient,
		ethClient:         ethClient,
		subscriptionState: subscriptionState,
		notifyEvents:      notifyEvents,
		logger:            logger.SetComponent(chainBlockEventsUseCaseComponent),
	}
}

func (uc *chainBlockEventsUC) Execute(ctx context.Context, chainUUID string, blockNumber uint64) error {
	logger := uc.logger.WithField("chain", chainUUID).WithField("block", blockNumber)

	addrs, err := uc.subscriptionState.ListAddressesPerChainUUID(ctx, chainUUID)
	if err != nil {
		errMsg := "failed to retrieve chain subscription addresses"
		uc.logger.WithError(err).Error(errMsg)
		return errors.FromError(err).SetMessage(errMsg)
	}

	for idx := range addrs {
		err := uc.handleAddressEvents(ctx, chainUUID, addrs[idx], blockNumber, logger)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uc *chainBlockEventsUC) handleAddressEvents(ctx context.Context, chainUUID string, addr ethcommon.Address, blockNumber uint64, logger *log.Logger) error {
	logger = logger.WithField("addr", addr.String())

	proxyURL := uc.proxyClient.ChainProxyURL(chainUUID)
	eventLogs, err := uc.ethClient.FilterLogs(ctx, proxyURL, []ethcommon.Address{addr}, new(big.Int).SetUint64(blockNumber),
		new(big.Int).SetUint64(blockNumber))
	if err != nil {
		logger.WithError(err).Error("failed to query filtered logs")
		return err
	}

	err = uc.notifyEvents.Execute(ctx, chainUUID, addr, eventLogs)
	if err != nil {
		return err
	}

	return nil
}
