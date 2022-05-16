package usecases

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

//go:generate mockgen -source=event_streams.go -destination=mocks/event_streams.go -package=mocks

type EventStreamsUseCases interface {
	Get() GetEventStreamUseCase
	Create() CreateEventStreamUseCase
	Update() UpdateEventStreamUseCase
	Search() SearchEventStreamsUseCase
	NotifyTransaction() NotifyTransactionUseCase
	NotifyContractEvents() NotifyContractEventsUseCase
	Delete() DeleteEventStreamUseCase
}

type GetEventStreamUseCase interface {
	Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) (*entities.EventStream, error)
}

type CreateEventStreamUseCase interface {
	Execute(ctx context.Context, eventStream *entities.EventStream, chainName string, userInfo *multitenancy.UserInfo) (*entities.EventStream, error)
}

type UpdateEventStreamUseCase interface {
	Execute(ctx context.Context, eventStream *entities.EventStream, userInfo *multitenancy.UserInfo) (*entities.EventStream, error)
}

type SearchEventStreamsUseCase interface {
	Execute(ctx context.Context, filters *entities.EventStreamFilters, userInfo *multitenancy.UserInfo) ([]*entities.EventStream, error)
}

type NotifyTransactionUseCase interface {
	Execute(ctx context.Context, job *entities.Job, errStr string, userInfo *multitenancy.UserInfo) error
}

type NotifyContractEventsUseCase interface {
	Execute(ctx context.Context, chainUUID string, address ethcommon.Address, eventLogs []ethtypes.Log, userInfo *multitenancy.UserInfo) error
}

type DeleteEventStreamUseCase interface {
	Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) error
}
