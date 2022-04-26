package usecases

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=event_streams.go -destination=mocks/event_streams.go -package=mocks

type EventStreamsUseCases interface {
	Get() GetEventStreamUseCase
	Create() CreateEventStreamUseCase
	Update() UpdateEventStreamUseCase
	Search() SearchEventStreamsUseCase
	NotifyTransaction() NotifyTransactionUseCase
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

type DeleteEventStreamUseCase interface {
	Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) error
}
