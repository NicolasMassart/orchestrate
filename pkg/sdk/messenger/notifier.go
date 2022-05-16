package messenger

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/notifier/service"
	"github.com/consensys/orchestrate/src/notifier/service/types"
)

func (c *ProducerClient) TransactionNotificationMessage(_ context.Context, eventStream *entities.EventStream, notif *entities.Notification, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicNotifier, service.TransactionMessageType, &types.TransactionMessageRequest{
		EventStream:  eventStream,
		Notification: notif,
	}, notif.SourceUUID, userInfo)
}

func (c *ProducerClient) ContractEventNotificationMessage(_ context.Context, eventStream *entities.EventStream, notif *entities.Notification, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicNotifier, service.ContractEventMessageType, &types.ContractEventMessageRequest{
		EventStream:  eventStream,
		Notification: notif,
	}, notif.SourceUUID, userInfo)
}
