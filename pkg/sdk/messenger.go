package sdk

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=messenger.go -destination=mock/messenger.go -package=mock

type OrchestrateMessenger interface {
	MessengerAPI
	MessengerNotifier
	MessengerTxListener
	MessengerTxSender
}

// @TODO Use MessageRequest types as input for methods, same than Orchestrate HTTP client

type MessengerAPI interface {
	ContractEventLogsMessage(ctx context.Context, chainUUID string, logs []*ethereum.Log, userInfo *multitenancy.UserInfo) error
	JobUpdateMessage(ctx context.Context, jobUUID string, status entities.JobStatus, msg string, userInfo *multitenancy.UserInfo) error
	EventStreamSuspendMessage(ctx context.Context, eventStreamUUID string, userInfo *multitenancy.UserInfo) error
	NotificationAckMessage(ctx context.Context, notifUUID string, userInfo *multitenancy.UserInfo) error
}

type MessengerNotifier interface {
	TransactionNotificationMessage(ctx context.Context, eventStream *entities.EventStream, notif *entities.Notification, userInfo *multitenancy.UserInfo) error
	ContractEventNotificationMessage(ctx context.Context, eventStream *entities.EventStream, subscriptionUUID string, eventLogs []*ethereum.Log, userInfo *multitenancy.UserInfo) error
}

type MessengerTxListener interface {
	PendingJobMessage(ctx context.Context, job *entities.Job, userInfo *multitenancy.UserInfo) error
	CreateSubscriptionMessage(ctx context.Context, sub *entities.Subscription, userInfo *multitenancy.UserInfo) error
	UpdateSubscriptionMessage(ctx context.Context, sub *entities.Subscription, userInfo *multitenancy.UserInfo) error
	DeleteSubscriptionMessage(ctx context.Context, sub *entities.Subscription, userInfo *multitenancy.UserInfo) error
}

type MessengerTxSender interface {
	StartedJobMessage(ctx context.Context, job *entities.Job, userInfo *multitenancy.UserInfo) error
}
