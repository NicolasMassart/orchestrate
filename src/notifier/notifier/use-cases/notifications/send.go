package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/consensys/orchestrate/src/notifier/store"

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
	db            store.NotificationAgent
}

func NewSendUseCase(db store.NotificationAgent, kafkaProducer kafka.Producer, webhookClient *http.Client) usecases.SendNotificationUseCase {
	return &sendUseCase{
		db:            db,
		kafkaProducer: kafkaProducer,
		webhookClient: webhookClient,
		logger:        log.NewLogger().SetComponent(sendComponent),
	}
}

func (uc *sendUseCase) Execute(ctx context.Context, eventStream *entities.EventStream, notif *entities.Notification) error {
	logger := uc.logger.WithContext(log.WithFields(ctx, log.Field("notification", notif.UUID), log.Field("event_stream", eventStream.UUID)))

	if eventStream.Status == entities.EventStreamStatusSuspend {
		logger.Warn("event stream is suspended")
		return nil
	}

	var err error
	switch eventStream.Channel {
	case entities.EventStreamChannelKafka:
		err = uc.sendKafkaNotificationResponse(ctx, eventStream, notif)
	case entities.EventStreamChannelWebhook:
		err = uc.sendWebhookNotificationResponse(ctx, eventStream, notif)
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

func (uc *sendUseCase) sendWebhookNotificationResponse(ctx context.Context, eventStream *entities.EventStream, notif *entities.Notification) error {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(types.NewNotificationResponse(notif))
	if err != nil {
		return errors.EncodingError(err.Error())
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, eventStream.Webhook.URL, body)
	req.Header.Set("Content-Type", "application/json")
	for key, val := range eventStream.Webhook.Headers {
		req.Header.Set(key, val)
	}

	_, err = uc.webhookClient.Do(req)
	if err != nil {
		return errors.HTTPConnectionError(err.Error())
	}

	return nil
}

func (uc *sendUseCase) sendKafkaNotificationResponse(_ context.Context, eventStream *entities.EventStream, notif *entities.Notification) error {
	err := uc.kafkaProducer.Send(types.NewNotificationResponse(notif), eventStream.Kafka.Topic, eventStream.ChainUUID, nil)
	if err != nil {
		return errors.KafkaConnectionError("could not send notification response. %s", err.Error())
	}

	return nil
}
