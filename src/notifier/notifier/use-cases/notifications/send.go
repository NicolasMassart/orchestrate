package notifications

import (
	"context"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/infra/webhook"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/kafka"
	usecases "github.com/consensys/orchestrate/src/notifier/notifier/use-cases"
	"github.com/consensys/orchestrate/src/notifier/service/types"
)

const sendComponent = "use-cases.notifier.send"

type sendUseCase struct {
	logger          *log.Logger
	kafkaProducer   kafka.Producer
	webhookProducer webhook.Producer
	messenger       sdk.MessengerAPI
}

func NewSendUseCase(
	kafkaProducer kafka.Producer,
	webhookProducer webhook.Producer,
	messenger sdk.MessengerAPI,
) usecases.SendNotificationUseCase {
	return &sendUseCase{
		kafkaProducer:   kafkaProducer,
		webhookProducer: webhookProducer,
		messenger:       messenger,
		logger:          log.NewLogger().SetComponent(sendComponent),
	}
}

func (uc *sendUseCase) Execute(ctx context.Context, eventStream *entities.EventStream, notif *entities.Notification) error {
	logger := uc.logger.WithContext(log.WithFields(ctx, log.Field("notification", notif.UUID), log.Field("event_stream", eventStream.UUID)))
	userInfo := multitenancy.NewInternalAdminUser()

	var err error
	switch eventStream.Channel {
	case entities.EventStreamChannelKafka:
		err = uc.kafkaProducer.Send(types.NewNotificationResponse(notif), eventStream.Kafka.Topic, eventStream.ChainUUID, nil)
	case entities.EventStreamChannelWebhook:
		err = uc.webhookProducer.Send(ctx, eventStream.Webhook, types.NewNotificationResponse(notif))
	default:
		return errors.InvalidParameterError("invalid event stream channel")
	}
	if err != nil {
		logger.WithError(err).Warn("failed to send notification")

		err = uc.messenger.EventStreamSuspendMessage(ctx, eventStream.UUID, userInfo)
		if err != nil {
			errMessage := "failed to suspend event stream"
			logger.WithError(err).Error(errMessage)
			return errors.DependencyFailureError(errMessage)
		}

		return nil
	}

	err = uc.messenger.NotificationAckMessage(ctx, notif.UUID, userInfo)
	if err != nil {
		errMessage := "failed to acknowledge notification"
		logger.WithError(err).Error(errMessage)
		return errors.DependencyFailureError(errMessage)
	}

	logger.Info("transaction notification sent successfully")
	return nil
}
