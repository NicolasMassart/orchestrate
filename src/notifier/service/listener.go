package service

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	infra "github.com/consensys/orchestrate/src/infra/api"
	"github.com/consensys/orchestrate/src/notifier/service/types"

	"github.com/Shopify/sarama"
	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/errors"
	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	usecases "github.com/consensys/orchestrate/src/notifier/notifier/use-cases"
)

const (
	messageListenerComponent = "notifier.kafka-consumer"
)

type MessageConsumerHandler struct {
	useCases          usecases.UseCases
	maxRetries        int
	eventStreamClient client.EventStreamClient
	logger            *log.Logger
}

func NewMessageConsumerHandler(useCases usecases.UseCases, eventStreamClient client.EventStreamClient, maxRetries int) *MessageConsumerHandler {
	return &MessageConsumerHandler{
		useCases:          useCases,
		maxRetries:        maxRetries,
		eventStreamClient: eventStreamClient,
		logger:            log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (mch *MessageConsumerHandler) DecodeMessage(rawMsg *sarama.ConsumerMessage) (interface{}, error) {
	request := &types.NotificationMessage{}
	err := infra.UnmarshalBody(bytes.NewReader(rawMsg.Value), request)
	if err != nil {
		errMessage := "failed to decode notification request"
		return nil, errors.InvalidFormatError(errMessage)
	}

	return request, nil
}

func (mch *MessageConsumerHandler) ID() string {
	return messageListenerComponent
}

func (mch *MessageConsumerHandler) ProcessMsg(ctx context.Context, rawMsg *sarama.ConsumerMessage, decodedMsg interface{}) error {
	req := decodedMsg.(*types.NotificationMessage)
	logger := mch.logger.WithField("event_stream", req.EventStream.UUID)

	for _, h := range rawMsg.Headers {
		if string(h.Key) == authutils.UserInfoHeader {
			userInfo := &multitenancy.UserInfo{}
			_ = json.Unmarshal(h.Value, userInfo)
			ctx = multitenancy.WithUserInfo(ctx, userInfo)
		}
	}

	err := backoff.RetryNotify(
		func() error {
			err := mch.sendNotification(ctx, req)
			switch {
			// Exits if not errors
			case err == nil:
				return nil
			case err == context.DeadlineExceeded || err == context.Canceled:
				return backoff.Permanent(err)
			case ctx.Err() != nil:
				return backoff.Permanent(ctx.Err())
			case errors.IsConnectionError(err):
				return err
			}

			return nil
		},
		backoff.NewConstantBackOff(time.Second),
		func(err error, duration time.Duration) {
			logger.WithError(err).Warnf("error processing notification, retrying in %v...", duration)
		},
	)
	if err != nil {
		// TODO: suspend the event stream if maxRetries is reached for the given notification
		return err
	}

	return nil
}

func (mch *MessageConsumerHandler) sendNotification(ctx context.Context, req *types.NotificationMessage) error {
	var notif *entities.Notification
	var err error

	switch req.Type {
	case types.TransactionNotificationType:
		notif, err = mch.useCases.CreateTransaction().Execute(ctx, req.Job, req.Error)
	case types.ContractEventNotificationType:
		return errors.InvalidParameterError("not implemented yet")
	default:
		return errors.InvalidParameterError("notification type %s is not supported", req.Type)
	}
	if err != nil {
		return err
	}

	err = mch.useCases.Send().Execute(ctx, notif, req.EventStream)
	if err != nil {
		return err
	}

	return nil
}
