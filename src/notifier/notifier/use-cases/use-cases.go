package usecases

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=use-cases.go -destination=mocks/use-cases.go -package=mocks

type SendNotificationUseCase interface {
	Execute(ctx context.Context, eventStream *entities.EventStream, notif *entities.Notification) error
}
