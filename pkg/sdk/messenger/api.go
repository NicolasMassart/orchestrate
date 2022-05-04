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
	return c.sendMessage(listener.JobUpdateMessageType, &types.EventLogsMessageRequest{
		ChainUUID: chainUUID,
		EventLogs: logs,
	}, chainUUID, userInfo)
}

func (c *ProducerClient) JobUpdateMessage(_ context.Context, jobUUID string, status entities.JobStatus, msg string, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(listener.JobUpdateMessageType, &types.JobUpdateMessageRequest{
		JobUUID: jobUUID,
		Status:  status,
		Message: msg,
	}, jobUUID, userInfo)
}
