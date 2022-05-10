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

func (c *ProducerClient) EventStreamUpdateMessage(_ context.Context, eventStreamUUID string, status entities.EventStreamStatus, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicAPI, listener.UpdateEventStreamMessageType, &types.UpdateEventStreamRequestMessage{
		UUID:   eventStreamUUID,
		Status: status,
	}, eventStreamUUID, userInfo)
}

func (c *ProducerClient) NotificationUpdateMessage(_ context.Context, notifUUID string, status entities.NotificationStatus, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicAPI, listener.UpdateNotificationMessageType, &types.UpdateNotificationRequestMessage{
		UUID:   notifUUID,
		Status: status,
	}, notifUUID, userInfo)
}
