package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/kafka"
	usecases "github.com/consensys/orchestrate/src/notifier/notifier/use-cases"
	"github.com/consensys/orchestrate/src/notifier/service/types"
)

const sendComponent = "use-cases.notifier.send"

type sendUseCase struct {
	logger        *log.Logger
	kafkaProducer kafka.Producer
	webhookClient *http.Client
	messenger     sdk.MessengerAPI
}

func NewSendUseCase(
	kafkaProducer kafka.Producer,
	webhookClient *http.Client,
	messenger sdk.MessengerAPI,
) usecases.SendNotificationUseCase {
	return &sendUseCase{
		kafkaProducer: kafkaProducer,
		webhookClient: webhookClient,
		messenger:     messenger,
		logger:        log.NewLogger().SetComponent(sendComponent),
	}
}

func (uc *sendUseCase) Execute(ctx context.Context, eventStream *entities.EventStream, notif *entities.Notification) error {
	logger := uc.logger.WithContext(log.WithFields(ctx, log.Field("notification", notif.UUID), log.Field("event_stream", eventStream.UUID)))
	userInfo := multitenancy.NewInternalAdminUser()

	if eventStream.Status == entities.EventStreamStatusSuspend {
		logger.Warn("event stream is suspended")
		return nil
	}

	var err error
	switch eventStream.Channel {
	case entities.EventStreamChannelKafka:
		err = uc.kafkaProducer.Send(types.NewNotificationResponse(notif), eventStream.Kafka.Topic, eventStream.ChainUUID, nil)
	case entities.EventStreamChannelWebhook:
		err = uc.sendWebhookNotificationResponse(ctx, eventStream, notif)
	default:
		return errors.InvalidParameterError("invalid event stream channel")
	}
	if err != nil {
		logger.WithError(err).Warn("failed to send notification")

		err = uc.messenger.EventStreamUpdateMessage(ctx, eventStream.UUID, entities.EventStreamStatusSuspend, userInfo)
		if err != nil {
			errMessage := "failed to suspend event stream"
			logger.WithError(err).Error(errMessage)
			return errors.DependencyFailureError(errMessage)
		}
	}

	err = uc.messenger.NotificationUpdateMessage(ctx, notif.UUID, entities.NotificationStatusSent, userInfo)
	if err != nil {
		errMessage := "failed to acknowledge notification"
		logger.WithError(err).Error(errMessage)
		return errors.DependencyFailureError(errMessage)
	}

	logger.Info("transaction notification sent successfully")
	return nil
}

func (uc *sendUseCase) sendWebhookNotificationResponse(ctx context.Context, eventStream *entities.EventStream, notif *entities.Notification) error {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(types.NewNotificationResponse(notif))
	if err != nil {
		return err
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, eventStream.Webhook.URL, body)
	req.Header.Set("Content-Type", "application/json")
	for key, val := range eventStream.Webhook.Headers {
		req.Header.Set(key, val)
	}

	_, err = uc.webhookClient.Do(req)
	if err != nil {
		return err
	}

	return nil
}
