package testutils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/consensys/orchestrate/src/infra/messenger/types"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	messenger "github.com/consensys/orchestrate/src/infra/messenger/kafka"
)

const (
	notifierMessageListenerComponent = "test.service.kafka-notification-consumer"
)

type notifierConsumerHandler struct {
	chanRegistry *utils.ChanRegistry
	logger       *log.Logger
}

var _ messenger.ConsumerMessageHandler = &notifierConsumerHandler{}

func newNotifierConsumerHandler(chanRegistry *utils.ChanRegistry) *notifierConsumerHandler {
	return &notifierConsumerHandler{
		chanRegistry: chanRegistry,
		logger:       log.NewLogger().SetComponent(notifierMessageListenerComponent),
	}
}

func (mch *notifierConsumerHandler) ProcessMsg(_ context.Context, rawMsg *sarama.ConsumerMessage, decodedMsg interface{}) error {
	notification := decodedMsg.(*types.NotificationResponse)
	logger := mch.logger.
		WithField("id", notification.UUID).
		WithField("source_uuid", notification.SourceUUID).
		WithField("type", notification.Type).
		WithField("topic", rawMsg.Topic)

	msgKey := keyGenOf(notification.SourceUUID, rawMsg.Topic)
	if !mch.chanRegistry.HasChan(msgKey) {
		mch.chanRegistry.Register(msgKey, make(chan interface{}, 1))
	}

	// Dispatch envelope
	err := mch.chanRegistry.Send(msgKey, notification)
	if err != nil {
		logger.WithError(err).Error("message dispatched with errors")
		return err
	}

	logger.Info("message has been processed")
	return nil
}

func (mch *notifierConsumerHandler) DecodeMessage(rawMsg *sarama.ConsumerMessage) (interface{}, error) {
	notif := &types.NotificationResponse{}
	err := json.Unmarshal(rawMsg.Value, notif)
	if err != nil {
		errMessage := "failed to decode notification message"
		return nil, errors.EncodingError(errMessage)
	}

	return notif, nil
}

func (mch *notifierConsumerHandler) ID() string {
	return notifierMessageListenerComponent
}

func keyGenOf(key, topic string) string {
	return fmt.Sprintf("%s/%s", topic, key)
}
