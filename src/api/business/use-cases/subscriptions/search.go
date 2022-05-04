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

const searchSubscriptionsComponent = "use-cases.search-subscription"

type searchUseCase struct {
	db     store.SubscriptionAgent
	logger *log.Logger
}

func NewSearchUseCase(db store.SubscriptionAgent) usecases.SearchSubscriptionUseCase {
	return &searchUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(searchSubscriptionsComponent),
	}
}

func (uc *searchUseCase) Execute(ctx context.Context, filters *entities.SubscriptionFilters, userInfo *multitenancy.UserInfo) ([]*entities.Subscription, error) {
	es, err := uc.db.Search(ctx, filters, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(searchSubscriptionsComponent)
	}

	uc.logger.WithContext(ctx).Debug("event streams found successfully")
	return es, nil
}
