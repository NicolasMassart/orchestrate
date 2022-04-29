package messenger

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	notifier "github.com/consensys/orchestrate/src/notifier/service/types"
)

//go:generate mockgen -source=producer.go -destination=mocks/producer.go -package=mocks

type Producer interface {
	SendJobMessage(topic string, job *entities.Job, partitionKey string, userInfo *multitenancy.UserInfo) error
	SendNotificationMessage(topic string, notif *notifier.NotificationMessage, partitionKey string, userInfo *multitenancy.UserInfo) error
	SendNotificationResponse(ctx context.Context, notif *entities.Notification, eventStream *entities.EventStream) error
	Checker() error
}
