// +build unit

package schedules

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSearchSchedules_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockScheduleDA := mocks.NewMockScheduleAgent(ctrl)
	mockJobDA := mocks.NewMockJobAgent(ctrl)

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewSearchSchedulesUseCase(mockDB)
	ctx := context.Background()

	t.Run("should execute use case successfully", func(t *testing.T) {
		scheduleEntity := testdata.FakeSchedule()
		expectedResponse := []*entities.Schedule{scheduleEntity}

		mockDB.EXPECT().Schedule().Return(mockScheduleDA).Times(1)
		mockDB.EXPECT().Job().Return(mockJobDA).Times(1)

		mockScheduleDA.EXPECT().
			FindAll(gomock.Any(), userInfo.AllowedTenants, userInfo.Username).
			Return([]*entities.Schedule{scheduleEntity}, nil)

		mockJobDA.EXPECT().
			FindOneByUUID(gomock.Any(), scheduleEntity.Jobs[0].UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(scheduleEntity.Jobs[0], nil)

		schedulesResponse, err := usecase.Execute(ctx, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, schedulesResponse)
	})

	t.Run("should fail with same error if FindAll fails for schedules", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		mockDB.EXPECT().Schedule().Return(mockScheduleDA).Times(1)

		mockScheduleDA.EXPECT().
			FindAll(gomock.Any(), userInfo.AllowedTenants, userInfo.Username).
			Return(nil, expectedErr)

		scheduleResponse, err := usecase.Execute(ctx, userInfo)

		assert.Nil(t, scheduleResponse)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(createScheduleComponent), err)
	})

	t.Run("should fail with same error if FindOne fails for jobs", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		scheduleEntity := testdata.FakeSchedule()

		mockDB.EXPECT().Schedule().Return(mockScheduleDA).Times(1)
		mockDB.EXPECT().Job().Return(mockJobDA).Times(1)

		mockScheduleDA.EXPECT().
			FindAll(gomock.Any(), userInfo.AllowedTenants, userInfo.Username).
			Return([]*entities.Schedule{scheduleEntity}, nil)

		mockJobDA.EXPECT().
			FindOneByUUID(gomock.Any(), scheduleEntity.Jobs[0].UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(scheduleEntity.Jobs[0], expectedErr)

		scheduleResponse, err := usecase.Execute(ctx, userInfo)

		assert.Nil(t, scheduleResponse)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(createScheduleComponent), err)
	})
}
