package usecases

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=event_streams.go -destination=mocks/event_streams.go -package=mocks

type EventStreamsUseCases interface {
	Create() CreateEventStreamUseCase
	Search() SearchEventStreamsUseCase
}

type CreateEventStreamUseCase interface {
	Execute(ctx context.Context, eventStream *entities.EventStream, userInfo *multitenancy.UserInfo) (*entities.EventStream, error)
}

type SearchEventStreamsUseCase interface {
	Execute(ctx context.Context, filters *entities.EventStreamFilters, userInfo *multitenancy.UserInfo) ([]*entities.EventStream, error)
}
