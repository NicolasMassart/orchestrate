package usecases

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=use-cases.go -destination=mocks/use-cases.go -package=mocks

type UseCases interface {
	Send() SendNotificationUseCase
	CreateTransaction() CreateTxNotificationUseCase
}

type CreateTxNotificationUseCase interface {
	Execute(ctx context.Context, job *entities.Job, errStr string) (*entities.Notification, error)
}

type SendNotificationUseCase interface {
	Execute(ctx context.Context, eventStream *entities.EventStream, notif *entities.Notification) error
}
