package usecases

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

//go:generate mockgen -source=subscriptions.go -destination=mocks/subscriptions.go -package=mocks

type CreatedSubscription interface {
	Execute(ctx context.Context, sub *entities.Subscription) error
}

type UpdatedSubscription interface {
	Execute(ctx context.Context, sub *entities.Subscription) error
}

type DeletedSubscription interface {
	Execute(ctx context.Context, subUUID string) error
}

type NotifySubscriptionEvents interface {
	Execute(ctx context.Context, chainUUID string, address ethcommon.Address, events []ethtypes.Log) error
}

type SubscriptionUseCases interface {
	CreatedSubscriptionUseCase() CreatedSubscription
	UpdatedSubscriptionUseCase() UpdatedSubscription
	DeletedSubscriptionJobUseCase() DeletedSubscription
	NotifySubscriptionEventsUseCase() NotifySubscriptionEvents
}
