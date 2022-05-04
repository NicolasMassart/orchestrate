package service

import (
	"bytes"
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/sdk"
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
	useCases usecases.UseCases,
	eventStreamClient sdk.EventStreamClient,
	maxRetries int,
) (*messenger.Consumer, error) {
	router := NewRouter(useCases, eventStreamClient, maxRetries)
	consumer, err := messenger.NewMessageConsumer(messageListenerComponent, cfg, topics)
	if err != nil {
		return nil, err
	}

	consumer.AppendHandler(TransactionMessageType, router.HandleTransactionReq)
	return consumer, nil
}

type Router struct {
	useCases          usecases.UseCases
	maxRetries        int
	eventStreamClient sdk.EventStreamClient
	logger            *log.Logger
}

func NewRouter(useCases usecases.UseCases, eventStreamClient sdk.EventStreamClient, maxRetries int) *Router {
	return &Router{
		useCases:          useCases,
		maxRetries:        maxRetries,
		eventStreamClient: eventStreamClient,
		logger:            log.NewLogger().SetComponent(messageListenerComponent),
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
			// Move to router method logic
			err = mch.sendTxNotification(ctx, req)
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

func (mch *Router) sendTxNotification(ctx context.Context, req *types.TransactionMessageRequest) error {
	// Validate request notifier
	notif, err := mch.useCases.CreateTransaction().Execute(ctx, req.Job, req.Error)
	if err != nil {
		return err
	}
	err = mch.useCases.Send().Execute(ctx, req.EventStream, notif)
	if err != nil {
		return err
	}

	return nil
}
