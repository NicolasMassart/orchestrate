package store

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=store.go -destination=mocks/mock.go -package=mocks

type NotificationAgent interface {
	Insert(ctx context.Context, notif *entities.Notification) (*entities.Notification, error)
	Update(ctx context.Context, notif *entities.Notification) (*entities.Notification, error)
}
