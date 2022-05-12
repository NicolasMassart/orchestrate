package usecases

import (
	"context"
)

//go:generate mockgen -source=notifications.go -destination=mocks/notifications.go -package=mocks

type NotificationsUseCases interface {
	Ack() AckNotificationUseCase
}

type AckNotificationUseCase interface {
	Execute(ctx context.Context, uuid string) error
}
