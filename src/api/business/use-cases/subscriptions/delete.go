package subscriptions

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
)

const deleteSubcriptionComponent = "use-cases.delete-subscription"

type deleteUseCase struct {
	db                  store.SubscriptionAgent
	txListenerMessenger sdk.MessengerTxListener
	logger              *log.Logger
}

func NewDeleteUseCase(db store.SubscriptionAgent, txListenerMessenger sdk.MessengerTxListener) usecases.DeleteSubscriptionUseCase {
	return &deleteUseCase{
		db:                  db,
		txListenerMessenger: txListenerMessenger,
		logger:              log.NewLogger().SetComponent(deleteSubcriptionComponent),
	}
}

func (uc *deleteUseCase) Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) error {
	ctx = log.WithFields(ctx, log.Field("subscription", uuid))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("deleting subscription")

	sub, err := uc.db.FindOneByUUID(ctx, uuid, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return errors.FromError(err).ExtendComponent(deleteSubcriptionComponent)
	}

	err = uc.db.Delete(ctx, uuid, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return errors.FromError(err).ExtendComponent(deleteSubcriptionComponent)
	}

	err = uc.txListenerMessenger.DeleteSubscriptionMessage(ctx, sub, userInfo)
	if err != nil {
		errMsg := "failed to send delete subscription message"
		uc.logger.WithError(err).Error(errMsg)
		return errors.DependencyFailureError(errMsg).ExtendComponent(createSubscriptionComponent)
	}

	logger.Info("subscription was deleted successfully")
	return nil
}
