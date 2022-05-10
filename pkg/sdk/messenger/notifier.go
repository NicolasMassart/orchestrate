package messenger

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/notifier/service"
	"github.com/consensys/orchestrate/src/notifier/service/types"
)

func (c *ProducerClient) TransactionNotificationMessage(_ context.Context, eventStream *entities.EventStream, notif *entities.Notification, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicTxNotifier, service.TransactionMessageType, &types.TransactionMessageRequest{
		EventStream:  eventStream,
		Notification: notif,
	}, notif.Job.PartitionKey(), userInfo)
}

func (c *ProducerClient) ContractEventNotificationMessage(_ context.Context, eventStream *entities.EventStream, subscriptionUUID string, eventLogs []*ethereum.Log, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicTxNotifier, service.ContractEventMessageType, &types.ContractEventMessageRequest{
		EventStream:      eventStream,
		SubscriptionUUID: subscriptionUUID,
		EventLogs:        eventLogs,
	}, subscriptionUUID, userInfo)
}
