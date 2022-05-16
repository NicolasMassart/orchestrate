package messenger

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/service/listener"
	"github.com/consensys/orchestrate/src/api/service/types"
)

func (c *ProducerClient) ContractEventLogsMessage(_ context.Context, req *types.EventLogsMessageRequest, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicAPI, listener.EventLogsMessageType, req, req.ChainUUID, userInfo)
}

func (c *ProducerClient) JobUpdateMessage(_ context.Context, req *types.JobUpdateMessageRequest, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicAPI, listener.UpdateJobMessageType, req, req.JobUUID, userInfo)
}

func (c *ProducerClient) EventStreamSuspendMessage(_ context.Context, eventStreamUUID string, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicAPI, listener.SuspendEventStreamMessageType, &types.SuspendEventStreamRequestMessage{
		UUID: eventStreamUUID,
	}, eventStreamUUID, userInfo)
}

func (c *ProducerClient) NotificationAckMessage(_ context.Context, notifUUID string, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicAPI, listener.AckNotificationMessageType, &types.AckNotificationRequestMessage{
		UUID: notifUUID,
	}, notifUUID, userInfo)
}
