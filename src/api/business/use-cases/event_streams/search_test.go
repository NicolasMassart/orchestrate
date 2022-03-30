// +build unit

package streams

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockEventStreamAgent(ctrl)

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	filter := &entities.EventStreamFilters{Names: []string{"eventStream1"}}

	usecase := NewSearchUseCase(mockDB)

	t.Run("should execute use case successfully", func(t *testing.T) {
		es := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().Search(gomock.Any(), filter, userInfo.AllowedTenants, userInfo.Username).Return([]*entities.EventStream{es}, nil)

		resp, err := usecase.Execute(ctx, filter, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, es, resp[0])
	})

	t.Run("should fail with same error if search fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		mockDB.EXPECT().Search(gomock.Any(), filter, userInfo.AllowedTenants, userInfo.Username).Return(nil, expectedErr)

		_, err := usecase.Execute(ctx, filter, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(searchEventStreamsComponent), err)
	})
}
