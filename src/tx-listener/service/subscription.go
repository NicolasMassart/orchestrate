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
	"github.com/consensys/orchestrate/src/tx-listener/service/types"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/sessions"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

var (
	SubscriptionMessageType entities.RequestMessageType = "subscription"
)

type SubscriptionHandler struct {
	subscriptionUCs  usecases.SubscriptionUseCases
	retryBackOff     backoff.BackOff
	chainSessionMngr sessions.ChainSessionManager
	logger           *log.Logger
}

func NewSubscriptionHandler(subscriptionUCs usecases.SubscriptionUseCases,
	chainSessionMngr sessions.ChainSessionManager,
	bck backoff.BackOff) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionUCs:  subscriptionUCs,
		chainSessionMngr: chainSessionMngr,
		retryBackOff:     bck,
		logger:           log.NewLogger().SetComponent(messageListenerComponent),
	}
}

func (mch *SubscriptionHandler) HandleSubscriptionMessage(ctx context.Context, msg *entities.Message) error {
	req := &types.SubscriptionMessageRequest{}
	err := api.UnmarshalBody(bytes.NewReader(msg.Body), req)
	if err != nil {
		return errors.InvalidFormatError("invalid pending job request type")
	}

	return backoff.RetryNotify(
		func() error {
			err := mch.processSubscription(ctx, req)
			switch {
			// Exits if not errors
			case err == nil:
				return nil
			case err == context.DeadlineExceeded || err == context.Canceled:
				return backoff.Permanent(ctx.Err())
			case ctx.Err() != nil:
				return backoff.Permanent(ctx.Err())
			case errors.IsConnectionError(err):
				return err
			default: // Remaining error types (err != nil)
				// @TODO Handle broken subscription
			}

			return nil
		},
		mch.retryBackOff,
		func(err error, duration time.Duration) {
			mch.logger.WithError(err).Warnf("error processing message, retrying in %v...", duration)
		},
	)
}

func (mch *SubscriptionHandler) processSubscription(ctx context.Context, req *types.SubscriptionMessageRequest) error {
	logger := mch.logger.WithField("subscription", req.Subscription.UUID).WithField("action", req.Action)

	switch req.Action {
	case types.CreateSubscriptionAction:
		err := mch.subscriptionUCs.CreatedSubscriptionUseCase().Execute(ctx, req.Subscription)
		if err != nil {
			logger.WithError(err).Error("failed to handle new subscription")
			return err
		}

		// @TODO Use full Chain object to avoid re fetching
		err = mch.chainSessionMngr.StartSession(ctx, req.Subscription.ChainUUID)
		if err != nil && !errors.IsAlreadyExistsError(err) {
			logger.WithError(err).Error("failed to start chain listening session")
			return err
		}
	case types.UpdateSubscriptionAction:
		err := mch.subscriptionUCs.UpdatedSubscriptionUseCase().Execute(ctx, req.Subscription)
		if err != nil {
			logger.WithError(err).Error("failed to handle updated subscription")
			return err
		}
	case types.DeleteSubscriptionAction:
		err := mch.subscriptionUCs.DeletedSubscriptionJobUseCase().Execute(ctx, req.Subscription.UUID)
		if err != nil {
			logger.WithError(err).Error("failed to handle deleted subscription")
			return err
		}
	}

	return nil
}
