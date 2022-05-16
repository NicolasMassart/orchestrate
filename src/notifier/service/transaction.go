package service

import (
	"bytes"
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/api"
	usecases "github.com/consensys/orchestrate/src/notifier/notifier/use-cases"
	"github.com/consensys/orchestrate/src/notifier/service/types"
)

var TransactionMessageType entities.RequestMessageType = "transaction_notification"

type TransactionHandler struct {
	sendUC     usecases.SendNotificationUseCase
	maxRetries int
	logger     *log.Logger
}

func NewTransactionHandler(sendUC usecases.SendNotificationUseCase, maxRetries int) *TransactionHandler {
	return &TransactionHandler{
		sendUC:     sendUC,
		maxRetries: maxRetries,
		logger:     log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (mch *TransactionHandler) HandleTransactionReq(ctx context.Context, msg *entities.Message) error {
	req := &types.TransactionMessageRequest{}
	err := api.UnmarshalBody(bytes.NewReader(msg.Body), req)
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
		return err
	}

	return nil
}
