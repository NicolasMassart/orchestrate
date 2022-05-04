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
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const createSubscriptionComponent = "use-cases.create-subscription"

type createUseCase struct {
	db                  store.SubscriptionAgent
	searchChainsUC      usecases.SearchChainsUseCase
	getContractUC       usecases.GetContractUseCase
	searchEventStreamUC usecases.SearchEventStreamsUseCase
	txListenerMessenger sdk.MessengerTxListener
	logger              *log.Logger
}

func NewCreateUseCase(db store.SubscriptionAgent,
	searchChainsUC usecases.SearchChainsUseCase,
	getContractUC usecases.GetContractUseCase,
	searchEventStreamUC usecases.SearchEventStreamsUseCase,
	txListenerMessenger sdk.MessengerTxListener,
) usecases.CreateSubscriptionUseCase {
	return &createUseCase{
		db:                  db,
		searchChainsUC:      searchChainsUC,
		getContractUC:       getContractUC,
		searchEventStreamUC: searchEventStreamUC,
		txListenerMessenger: txListenerMessenger,
		logger:              log.NewLogger().SetComponent(createSubscriptionComponent),
	}
}

func (uc *createUseCase) Execute(ctx context.Context, subscription *entities.Subscription, chainName,
	eventStreamName string, userInfo *multitenancy.UserInfo) (*entities.Subscription, error) {
	ctx = log.WithFields(ctx, log.Field("address", subscription.Address))
	logger := uc.logger.WithContext(ctx)

	logger.Debug("creating new subscription")

	chains, err := uc.searchChainsUC.Execute(ctx, &entities.ChainFilters{Names: []string{chainName}}, userInfo)
	if err != nil {
		return nil, err
	}
	if len(chains) == 0 {
		errMessage := fmt.Sprintf("chain '%s' does not exist", chainName)
		uc.logger.WithContext(ctx).Error(errMessage)
		return nil, errors.InvalidParameterError(errMessage).ExtendComponent(createSubscriptionComponent)
	}
	subscription.ChainUUID = chains[0].UUID

	subscriptions, err := uc.db.Search(ctx,
		&entities.SubscriptionFilters{Addresses: []ethcommon.Address{subscription.Address},
			ChainUUID: subscription.ChainUUID, TenantID: userInfo.TenantID},
		userInfo.AllowedTenants,
		userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(createSubscriptionComponent)
	}
	if len(subscriptions) > 0 {
		errMsg := "subscription with same address and chain already exists"
		logger.Error(errMsg)
		return nil, errors.AlreadyExistsError(errMsg).ExtendComponent(createSubscriptionComponent)
	}

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

	contract, err := uc.getContractUC.Execute(ctx, subscription.ContractName, subscription.ContractTag)
	if err != nil {
		return nil, errors.FromError(err).SetComponent(createSubscriptionComponent)
	} else if contract == nil {
		errMessage := fmt.Sprintf("contract '%s:%s' does not exist", subscription.ContractName, subscription.ContractTag)
		uc.logger.WithContext(ctx).Error(errMessage)
		return nil, errors.InvalidParameterError(errMessage).ExtendComponent(createSubscriptionComponent)
	}

	subscription.TenantID = userInfo.TenantID
	subscription.OwnerID = userInfo.Username
	sub, err := uc.db.Insert(ctx, subscription)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(createSubscriptionComponent)
	}

	err = uc.txListenerMessenger.CreateSubscriptionMessage(ctx, sub, userInfo)
	if err != nil {
		errMsg := "failed to send create subscription message"
		uc.logger.WithError(err).Error(errMsg)
		return nil, errors.DependencyFailureError(errMsg).ExtendComponent(createSubscriptionComponent)
	}

	logger.WithField("subscription", sub.UUID).Info("subscription created successfully")
	return sub, nil
}
