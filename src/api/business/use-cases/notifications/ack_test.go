//go:build unit
// +build unit

package notifications

import (
	"context"
	"github.com/gofrs/uuid"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAck_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockNotificationAgent(ctrl)

	usecase := NewAckUseCase(mockDB)

	t.Run("should execute use case successfully", func(t *testing.T) {
		expectedUuid := uuid.Must(uuid.NewV4()).String()
		expectedNotif := &entities.Notification{UUID: expectedUuid, Status: entities.NotificationStatusSent}

		mockDB.EXPECT().Update(gomock.Any(), expectedNotif).Return(expectedNotif, nil)

		err := usecase.Execute(context.Background(), expectedUuid)

		assert.NoError(t, err)
	})

	t.Run("should fail with same error if fails to update", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		expectedUuid := uuid.Must(uuid.NewV4()).String()

		mockDB.EXPECT().Update(gomock.Any(), &entities.Notification{UUID: expectedUuid, Status: entities.NotificationStatusSent}).Return(&entities.Notification{}, expectedErr)

		err := usecase.Execute(context.Background(), expectedUuid)

		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(ackNotifComponent), err)
	})
}
