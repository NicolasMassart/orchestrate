// +build unit

package streams

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockEventStreamAgent(ctrl)

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewUpdateUseCase(mockDB)

	t.Run("should update event stream successfully", func(t *testing.T) {
		eventStream := testdata.FakeWebhookEventStream()
		eventStream.TenantID = userInfo.TenantID
		eventStream.OwnerID = userInfo.Username

		mockDB.EXPECT().Search(
			gomock.Any(),
			&entities.EventStreamFilters{Names: []string{eventStream.Name}, TenantID: userInfo.TenantID},
			userInfo.AllowedTenants,
			userInfo.Username,
		).Return([]*entities.EventStream{}, nil)
		mockDB.EXPECT().Update(gomock.Any(), eventStream, userInfo.AllowedTenants, userInfo.Username).Return(eventStream, nil)
		mockDB.EXPECT().FindOneByUUID(gomock.Any(), eventStream.UUID, userInfo.AllowedTenants, userInfo.Username).Return(eventStream, nil)

		resp, err := usecase.Execute(ctx, eventStream, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, resp, eventStream)
	})

	t.Run("should fail with same error if search event streams fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		eventStream := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)

		_, err := usecase.Execute(ctx, eventStream, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateEventStreamComponent), err)
	})

	t.Run("should fail with AlreadyExistsError if search event streams returns values", func(t *testing.T) {
		eventStream := testdata.FakeWebhookEventStream()
		foundEventStreamEntity := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*entities.EventStream{foundEventStreamEntity}, nil)

		_, err := usecase.Execute(ctx, eventStream, userInfo)
		assert.Error(t, err)
		assert.True(t, errors.IsAlreadyExistsError(err))
	})

	t.Run("should fail with same error if cannot update event stream", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		eventStream := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*entities.EventStream{}, nil)
		mockDB.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)

		_, err := usecase.Execute(ctx, eventStream, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateEventStreamComponent), err)
	})

	t.Run("should fail with same error if cannot get event stream", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		eventStream := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*entities.EventStream{}, nil)
		mockDB.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(eventStream, nil)
		mockDB.EXPECT().FindOneByUUID(gomock.Any(), eventStream.UUID, userInfo.AllowedTenants, userInfo.Username).Return(nil, expectedErr)

		_, err := usecase.Execute(ctx, eventStream, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateEventStreamComponent), err)
	})
}
