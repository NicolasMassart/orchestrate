package listener

import (
	"bytes"
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/infra/api"
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	messenger "github.com/consensys/orchestrate/src/infra/messenger/kafka"
)

const (
	messageListenerComponent = "api.kafka-consumer"
)

func NewMessageConsumer(cfg *kafka.Config,
	topics []string,
	notifySubscriptionUC usecases.NotifySubscriptionEventUseCase,
	updateJobUC usecases.UpdateJobUseCase,
) (*messenger.Consumer, error) {
	consumer, err := messenger.NewMessageConsumer(messageListenerComponent, cfg, topics)
	if err != nil {
		return nil, err
	}
	router := NewRouter(notifySubscriptionUC, updateJobUC)
	consumer.AppendHandler(EventLogsMessageType, router.HandleEventLogs)
	consumer.AppendHandler(JobUpdateMessageType, router.HandleJobUpdate)
	return consumer, nil
}

type Router struct {
	notifySubscriptionUC usecases.NotifySubscriptionEventUseCase
	updateJobUC          usecases.UpdateJobUseCase
	logger               *log.Logger
}

func NewRouter(notifySubscriptionUC usecases.NotifySubscriptionEventUseCase,
	updateJobUC usecases.UpdateJobUseCase) *Router {
	return &Router{
		notifySubscriptionUC: notifySubscriptionUC,
		updateJobUC:          updateJobUC,
		logger:               log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (mch *Router) HandleEventLogs(ctx context.Context, rawReq []byte) error {
	req := &types.EventLogsMessageRequest{}
	err := api.UnmarshalBody(bytes.NewReader(rawReq), req)
	if err != nil {
		return errors.InvalidFormatError("invalid event logs request type")
	}

	logger := mch.logger.WithField("chain", req.ChainUUID).WithField("address", req.Address)
	logger.WithField("length", len(req.EventLogs)).Debug("handling event logs")

	userInfo := multitenancy.UserInfoValue(ctx)

	err = mch.notifySubscriptionUC.Execute(ctx, req.ChainUUID, req.Address, req.EventLogs, userInfo)
	return err
}

func (mch *Router) HandleJobUpdate(ctx context.Context, rawReq []byte) error {
	req := &types.JobUpdateMessageRequest{}
	err := api.UnmarshalBody(bytes.NewReader(rawReq), req)
	if err != nil {
		return errors.InvalidFormatError("invalid job update request type")
	}
	logger := mch.logger.WithField("job", req.JobUUID)
	logger.WithField("status", req.Status).Debug("updating job")

	userInfo := multitenancy.UserInfoValue(ctx)

	_, err = mch.updateJobUC.Execute(ctx, req.ToJobEntity(), req.Status, req.Message, userInfo)
	return err
}
