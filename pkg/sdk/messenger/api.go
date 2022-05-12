package messenger

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/src/api/service/listener"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
)

func (c *ProducerClient) ContractEventLogsMessage(_ context.Context, chainUUID string, logs []*ethereum.Log, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicAPI, listener.UpdateJobMessageType, &types.EventLogsMessageRequest{
		ChainUUID: chainUUID,
		EventLogs: logs,
	}, chainUUID, userInfo)
}

func (c *ProducerClient) JobUpdateMessage(_ context.Context, jobUUID string, status entities.JobStatus, msg string, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicAPI, listener.UpdateJobMessageType, &types.JobUpdateMessageRequest{
		JobUUID: jobUUID,
		Status:  status,
		Message: msg,
	}, jobUUID, userInfo)
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
