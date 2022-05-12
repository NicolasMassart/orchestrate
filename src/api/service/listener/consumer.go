package listener

import (
	"bytes"
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/api"
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	messenger "github.com/consensys/orchestrate/src/infra/messenger/kafka"
)

const messageListenerComponent = "api.kafka-consumer"

func NewMessageConsumer(
	cfg *kafka.Config,
	topics []string,
	notifySubscriptionUC usecases.NotifySubscriptionEventUseCase,
	updateJobUC usecases.UpdateJobUseCase,
	ackNotifUC usecases.AckNotificationUseCase,
	updateEventStreamUC usecases.UpdateEventStreamUseCase,
) (*messenger.Consumer, error) {
	consumer, err := messenger.NewMessageConsumer(messageListenerComponent, cfg, topics)
	if err != nil {
		return nil, err
	}

	r := newRouter(notifySubscriptionUC, updateJobUC, ackNotifUC, updateEventStreamUC)

	consumer.AppendHandler(EventLogsMessageType, r.HandleEventLogs)
	consumer.AppendHandler(UpdateJobMessageType, r.HandleJobUpdate)
	consumer.AppendHandler(AckNotificationMessageType, r.HandleNotificationAck)
	consumer.AppendHandler(SuspendEventStreamMessageType, r.HandleEventStreamSuspend)

	return consumer, nil
}

type router struct {
	notifySubscriptionUC usecases.NotifySubscriptionEventUseCase
	updateJobUC          usecases.UpdateJobUseCase
	ackNotifUC           usecases.AckNotificationUseCase
	updateEventStreamUC  usecases.UpdateEventStreamUseCase
}

func newRouter(
	notifySubscriptionUC usecases.NotifySubscriptionEventUseCase,
	updateJobUC usecases.UpdateJobUseCase,
	ackNotifUC usecases.AckNotificationUseCase,
	updateEventStreamUC usecases.UpdateEventStreamUseCase,
) *router {
	return &router{
		notifySubscriptionUC: notifySubscriptionUC,
		updateJobUC:          updateJobUC,
		ackNotifUC:           ackNotifUC,
		updateEventStreamUC:  updateEventStreamUC,
	}
}

func (r *router) HandleEventLogs(ctx context.Context, rawReq []byte) error {
	req := &types.EventLogsMessageRequest{}
	err := api.UnmarshalBody(bytes.NewReader(rawReq), req)
	if err != nil {
		return errors.InvalidFormatError("invalid event logs request type")
	}

	userInfo := multitenancy.UserInfoValue(ctx)

	err = r.notifySubscriptionUC.Execute(ctx, req.ChainUUID, req.Address, req.EventLogs, userInfo)
	return err
}

func (r *router) HandleJobUpdate(ctx context.Context, rawReq []byte) error {
	req := &types.JobUpdateMessageRequest{}
	err := api.UnmarshalBody(bytes.NewReader(rawReq), req)
	if err != nil {
		return errors.InvalidFormatError("invalid job update request type")
	}

	userInfo := multitenancy.UserInfoValue(ctx)

	_, err = r.updateJobUC.Execute(ctx, req.ToJobEntity(), req.Status, req.Message, userInfo)
	return err
}

func (r *router) HandleNotificationAck(ctx context.Context, rawReq []byte) error {
	req := &types.AckNotificationRequestMessage{}
	err := api.UnmarshalBody(bytes.NewReader(rawReq), req)
	if err != nil {
		return errors.InvalidFormatError("invalid notification ack request type")
	}

	return r.ackNotifUC.Execute(ctx, req.UUID)
}

func (r *router) HandleEventStreamSuspend(ctx context.Context, rawReq []byte) error {
	req := &types.SuspendEventStreamRequestMessage{}
	err := api.UnmarshalBody(bytes.NewReader(rawReq), req)
	if err != nil {
		return errors.InvalidFormatError("invalid event stream suspend request type")
	}

	userInfo := multitenancy.UserInfoValue(ctx)

	_, err = r.updateEventStreamUC.Execute(ctx, &entities.EventStream{
		UUID:   req.UUID,
		Status: entities.EventStreamStatusSuspend,
	}, userInfo)

	return err
}
