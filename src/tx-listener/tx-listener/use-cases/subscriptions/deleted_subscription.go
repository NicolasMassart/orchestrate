package subscriptions

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const deletedSubscriptionUseCaseComponent = "tx-listener.use-case.deleted-subscription"

type deletedSubscriptionUC struct {
	subscriptionState store.Subscriptions
	logger            *log.Logger
}

func DeletedSubscriptionUseCase(subscriptionState store.Subscriptions,
	logger *log.Logger,
) usecases.DeletedSubscription {
	return &deletedSubscriptionUC{
		subscriptionState: subscriptionState,
		logger:            logger.SetComponent(deletedSubscriptionUseCaseComponent),
	}
}

func (uc *deletedSubscriptionUC) Execute(ctx context.Context, subUUID string) error {
	logger := uc.logger.WithField("subscription", subUUID)

	logger.Debug("handling deleted subscription")

	err := uc.subscriptionState.Remove(ctx, subUUID)
	if err != nil {
		logger.WithError(err).Error("failed to delete subscription")
		return err
	}

	return nil
}
