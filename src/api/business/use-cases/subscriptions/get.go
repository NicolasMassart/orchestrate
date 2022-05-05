package subscriptions

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
)

const getSubscriptionComponent = "use-cases.get-subscription"

type getUseCase struct {
	db     store.SubscriptionAgent
	logger *log.Logger
}

func NewGetUseCase(db store.SubscriptionAgent) usecases.GetSubscriptionUseCase {
	return &getUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(getSubscriptionComponent),
	}
}

func (uc *getUseCase) Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) (*entities.Subscription, error) {
	ctx = log.WithFields(ctx, log.Field("subscription", uuid))
	logger := uc.logger.WithContext(ctx)

	e, err := uc.db.FindOneByUUID(ctx, uuid, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(getSubscriptionComponent)
	}

	logger.WithField("address", e.ContractAddress.String()).WithField("chain", e.ChainUUID).
		Debug("subscription found successfully")
	return e, nil
}
