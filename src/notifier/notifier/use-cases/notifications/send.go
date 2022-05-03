package notifications

import (
	"context"

	"github.com/consensys/orchestrate/src/notifier/store"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/messenger"
	usecases "github.com/consensys/orchestrate/src/notifier/notifier/use-cases"
)

const sendComponent = "use-cases.notifier.send"

type sendUseCase struct {
	logger          *log.Logger
	kafkaNotifier   messenger.Producer
	webhookNotifier messenger.Producer
	db              store.NotificationAgent
}

func NewSendUseCase(db store.NotificationAgent, kafkaNotifier, webhookNotifier messenger.Producer) usecases.SendNotificationUseCase {
	return &sendUseCase{
		db:              db,
		kafkaNotifier:   kafkaNotifier,
		webhookNotifier: webhookNotifier,
		logger:          log.NewLogger().SetComponent(sendComponent),
	}
}

func (uc *sendUseCase) Execute(ctx context.Context, notif *entities.Notification, eventStream *entities.EventStream) error {
	logger := uc.logger.WithContext(log.WithFields(ctx, log.Field("notification", notif.UUID), log.Field("event_stream", eventStream.UUID)))

	if eventStream.Status == entities.EventStreamStatusSuspend {
		logger.Warn("event stream is suspended")
		return nil
	}

	var err error
	switch eventStream.Channel {
	case entities.EventStreamChannelKafka:
		err = uc.kafkaNotifier.SendNotificationResponse(ctx, notif, eventStream)
	case entities.EventStreamChannelWebhook:
		err = uc.webhookNotifier.SendNotificationResponse(ctx, notif, eventStream)
	default:
		return errors.InvalidParameterError("invalid event stream channel")
	}
	if err != nil {
		errMessage := "failed to send notification to the specified event stream"
		logger.WithError(err).Error(errMessage)
		return errors.DependencyFailureError(errMessage)
	}

	notif.Status = entities.NotificationStatusSent
	_, err = uc.db.Update(ctx, notif)
	if err != nil {
		return errors.FromError(err).ExtendComponent(createTxComponent)
	}

	logger.Info("transaction notification sent successfully")
	return nil
}
