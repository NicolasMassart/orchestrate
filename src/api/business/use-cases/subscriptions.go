package usecases

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -source=subscriptions.go -destination=mocks/subscriptions.go -package=mocks

type SubscriptionUseCases interface {
	Get() GetSubscriptionUseCase
	Create() CreateSubscriptionUseCase
	Update() UpdateSubscriptionUseCase
	Search() SearchSubscriptionUseCase
	NotifySubscription() NotifySubscriptionEventUseCase
	Delete() DeleteSubscriptionUseCase
}

type GetSubscriptionUseCase interface {
	Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) (*entities.Subscription, error)
}

type CreateSubscriptionUseCase interface {
	Execute(ctx context.Context, Subscription *entities.Subscription, chainName, eventStreamName string, userInfo *multitenancy.UserInfo) (*entities.Subscription, error)
}

type UpdateSubscriptionUseCase interface {
	Execute(ctx context.Context, Subscription *entities.Subscription, eventStreamName string, userInfo *multitenancy.UserInfo) (*entities.Subscription, error)
}

type SearchSubscriptionUseCase interface {
	Execute(ctx context.Context, filters *entities.SubscriptionFilters, userInfo *multitenancy.UserInfo) ([]*entities.Subscription, error)
}

type NotifySubscriptionEventUseCase interface {
	Execute(ctx context.Context, chainUUID string, address ethcommon.Address, eventLogs []*ethereum.Log, userInfo *multitenancy.UserInfo) error
}

type DeleteSubscriptionUseCase interface {
	Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) error
}
