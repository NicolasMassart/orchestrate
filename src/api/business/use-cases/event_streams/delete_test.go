// +build unit

package streams

import (
	"context"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockEventStreamAgent(ctrl)

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	uuid := "uuid"

	usecase := NewDeleteUseCase(mockDB)

	t.Run("should delete event stream successfully", func(t *testing.T) {
		mockDB.EXPECT().FindOneByUUID(gomock.Any(), uuid, userInfo.AllowedTenants, userInfo.Username).Return(testdata.FakeWebhookEventStream(), nil)
		mockDB.EXPECT().Delete(gomock.Any(), uuid, userInfo.AllowedTenants, userInfo.Username).Return(nil)

		err := usecase.Execute(ctx, uuid, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should fail with same error if cannot get event stream", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		mockDB.EXPECT().FindOneByUUID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)

		err := usecase.Execute(ctx, uuid, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(deleteEventStreamComponent), err)
	})

	t.Run("should fail with same error if cannot get event stream", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		mockDB.EXPECT().FindOneByUUID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(testdata.FakeWebhookEventStream(), nil)
		mockDB.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

		err := usecase.Execute(ctx, uuid, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(deleteEventStreamComponent), err)
	})
}
