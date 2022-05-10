package messenger

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-sender/service"
	"github.com/consensys/orchestrate/src/tx-sender/service/types"
)

func (c *ProducerClient) StartedJobMessage(_ context.Context, job *entities.Job, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(c.cfg.TopicTxSender, service.StartedJobMessageType, &types.StartedJobReq{
		Job: job,
	}, job.PartitionKey(), userInfo)
}
