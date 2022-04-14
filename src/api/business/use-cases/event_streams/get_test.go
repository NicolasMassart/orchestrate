// +build unit

package streams

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockEventStreamAgent(ctrl)

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	uuid := "uuid"

	usecase := NewGetUseCase(mockDB)

	t.Run("should get event stream successfully", func(t *testing.T) {
		eventStream := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().FindOneByUUID(gomock.Any(), uuid, userInfo.AllowedTenants, userInfo.Username).Return(eventStream, nil)

		resp, err := usecase.Execute(ctx, uuid, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, resp, eventStream)
	})

	t.Run("should fail with same error if cannot get event stream", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		mockDB.EXPECT().FindOneByUUID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)

		_, err := usecase.Execute(ctx, uuid, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(getEventStreamComponent), err)
	})
}
