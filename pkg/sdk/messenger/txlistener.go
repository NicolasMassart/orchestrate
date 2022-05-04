package messenger

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/service"
	"github.com/consensys/orchestrate/src/tx-listener/service/types"
)

func (c *ProducerClient) PendingJobMessage(_ context.Context, job *entities.Job, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(service.PendingJobMessageType, &types.PendingJobMessageRequest{
		Job: job,
	}, job.ChainUUID, userInfo)
}

func (c *ProducerClient) CreateSubscriptionMessage(_ context.Context, sub *entities.Subscription, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(service.SubscriptionMessageType, &types.SubscriptionMessageRequest{
		Action:       types.CreateSubscriptionAction,
		Subscription: sub,
	}, sub.ChainUUID, userInfo)
}

func (c *ProducerClient) UpdateSubscriptionMessage(_ context.Context, sub *entities.Subscription, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(service.SubscriptionMessageType, &types.SubscriptionMessageRequest{
		Action:       types.UpdateSubscriptionAction,
		Subscription: sub,
	}, sub.ChainUUID, userInfo)
}

func (c *ProducerClient) DeleteSubscriptionMessage(_ context.Context, sub *entities.Subscription, userInfo *multitenancy.UserInfo) error {
	return c.sendMessage(service.SubscriptionMessageType, &types.SubscriptionMessageRequest{
		Action:       types.DeleteSubscriptionAction,
		Subscription: sub,
	}, sub.ChainUUID, userInfo)
}
