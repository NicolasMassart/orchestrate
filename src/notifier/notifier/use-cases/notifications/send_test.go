//go:build unit
// +build unit

package notifications

import (
	"context"
	"github.com/consensys/orchestrate/pkg/sdk/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mocks2 "github.com/consensys/orchestrate/src/infra/kafka/mocks"
	"github.com/consensys/orchestrate/src/infra/webhook/mocks"
	"github.com/consensys/orchestrate/src/notifier/service/types"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSend_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKafkaProducer := mocks2.NewMockProducer(ctrl)
	mockMessenger := mock.NewMockMessengerAPI(ctrl)
	mockWebhookProducer := mocks.NewMockProducer(ctrl)

	userInfo := multitenancy.NewInternalAdminUser()

	usecase := NewSendUseCase(mockKafkaProducer, mockWebhookProducer, mockMessenger)

	t.Run("should execute use case successfully: webhook ", func(t *testing.T) {
		es := testdata.FakeWebhookEventStream()
		notif := testdata.FakeNotification()

		mockWebhookProducer.EXPECT().Send(gomock.Any(), es.Webhook, types.NewNotificationResponse(notif)).Return(nil)
		mockMessenger.EXPECT().NotificationAckMessage(gomock.Any(), notif.UUID, userInfo).Return(nil)

		err := usecase.Execute(context.Background(), es, notif)

		assert.NoError(t, err)
	})

	t.Run("should execute use case successfully: kafka ", func(t *testing.T) {
		es := testdata.FakeKafkaEventStream()
		notif := testdata.FakeNotification()

		mockKafkaProducer.EXPECT().Send(types.NewNotificationResponse(notif), es.Kafka.Topic, es.ChainUUID, nil).Return(nil)
		mockMessenger.EXPECT().NotificationAckMessage(gomock.Any(), notif.UUID, userInfo).Return(nil)

		err := usecase.Execute(context.Background(), es, notif)

		assert.NoError(t, err)
	})

	t.Run("should suspend event stream if it fails to send notification", func(t *testing.T) {
		es := testdata.FakeKafkaEventStream()
		notif := testdata.FakeNotification()

		mockKafkaProducer.EXPECT().Send(types.NewNotificationResponse(notif), es.Kafka.Topic, es.ChainUUID, nil).Return(errors.KafkaConnectionError("error"))
		mockMessenger.EXPECT().EventStreamSuspendMessage(gomock.Any(), es.UUID, userInfo).Return(nil)

		err := usecase.Execute(context.Background(), es, notif)

		assert.NoError(t, err)
	})

	t.Run("should fail with Dependency error if it fails to suspend", func(t *testing.T) {
		es := testdata.FakeKafkaEventStream()
		notif := testdata.FakeNotification()

		mockKafkaProducer.EXPECT().Send(types.NewNotificationResponse(notif), es.Kafka.Topic, es.ChainUUID, nil).Return(errors.KafkaConnectionError("error"))
		mockMessenger.EXPECT().EventStreamSuspendMessage(gomock.Any(), es.UUID, userInfo).Return(errors.KafkaConnectionError("error"))

		err := usecase.Execute(context.Background(), es, notif)

		assert.True(t, errors.IsDependencyFailureError(err))
	})

	t.Run("should fail with Dependency error if it fails to acknowledge", func(t *testing.T) {
		es := testdata.FakeKafkaEventStream()
		notif := testdata.FakeNotification()

		mockKafkaProducer.EXPECT().Send(types.NewNotificationResponse(notif), es.Kafka.Topic, es.ChainUUID, nil).Return(nil)
		mockMessenger.EXPECT().NotificationAckMessage(gomock.Any(), notif.UUID, userInfo).Return(errors.KafkaConnectionError("error"))

		err := usecase.Execute(context.Background(), es, notif)

		assert.True(t, errors.IsDependencyFailureError(err))
	})
}
