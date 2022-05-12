package builder

import (
	"github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/notifications"
	"github.com/consensys/orchestrate/src/api/store"
)

type notificationUseCases struct {
	ack usecases.AckNotificationUseCase
}

var _ usecases.NotificationsUseCases = &notificationUseCases{}

func NewNotificationUseCases(db store.NotificationAgent) *notificationUseCases {
	return &notificationUseCases{
		ack: notifications.NewAckUseCase(db),
	}
}

func (u *notificationUseCases) Ack() usecases.AckNotificationUseCase {
	return u.ack
}
