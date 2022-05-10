package service

import (
	"bytes"
	"context"
	"time"

	"github.com/consensys/orchestrate/src/infra/api"
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	messenger "github.com/consensys/orchestrate/src/infra/messenger/kafka"
	"github.com/consensys/orchestrate/src/notifier/service/types"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/notifier/notifier/use-cases"
)

const (
	messageListenerComponent = "notifier.kafka-consumer"
)

func NewMessageConsumer(cfg *kafka.Config,
	topics []string,
	sendUC usecases.SendNotificationUseCase,
	maxRetries int,
) (*messenger.Consumer, error) {
	router := NewRouter(sendUC, maxRetries)
	consumer, err := messenger.NewMessageConsumer(messageListenerComponent, cfg, topics)
	if err != nil {
		return nil, err
	}

	consumer.AppendHandler(TransactionMessageType, router.HandleTransactionReq)
	return consumer, nil
}

type Router struct {
	sendUC     usecases.SendNotificationUseCase
	maxRetries int
	logger     *log.Logger
}

func NewRouter(sendUC usecases.SendNotificationUseCase, maxRetries int) *Router {
	return &Router{
		sendUC:     sendUC,
		maxRetries: maxRetries,
		logger:     log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (mch *Router) HandleTransactionReq(ctx context.Context, rawReq []byte) error {
	req := &types.TransactionMessageRequest{}
	err := api.UnmarshalBody(bytes.NewReader(rawReq), req)
	if err != nil {
		return errors.InvalidFormatError("invalid transaction request type")
	}

	err = backoff.RetryNotify(
		func() error {
			err = mch.sendUC.Execute(ctx, req.EventStream, req.Notification)
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
			mch.logger.WithError(err).Warnf("error processing notification, retrying in %v...", duration)
		},
	)
	if err != nil {
		// TODO: suspend the event stream if maxRetries is reached for the given notification
		return err
	}

	return nil
}
