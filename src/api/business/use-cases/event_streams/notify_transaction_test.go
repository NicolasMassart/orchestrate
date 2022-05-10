//go:build unit
// +build unit

package streams

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/mock"
	mocks3 "github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNotifyTx(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockEventStream := mocks.NewMockEventStreamAgent(ctrl)
	mockNotification := mocks.NewMockNotificationAgent(ctrl)
	messenger := mock.NewMockMessengerNotifier(ctrl)
	searchContractsUC := mocks3.NewMockSearchContractUseCase(ctrl)
	decodeLogUC := mocks3.NewMockDecodeEventLogUseCase(ctrl)

	mockDB.EXPECT().EventStream().Return(mockEventStream).AnyTimes()
	mockDB.EXPECT().Notification().Return(mockNotification).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	errStr := "error"

	usecase := NewNotifyTransactionUseCase(mockDB, searchContractsUC, decodeLogUC, messenger)

	t.Run("should execute use case successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Status = entities.StatusFailed
		eventStream := testdata.FakeWebhookEventStream()
		expectedNotif := &entities.Notification{
			SourceUUID: job.ScheduleUUID,
			SourceType: entities.NotificationSourceTypeJob,
			Status:     entities.NotificationStatusPending,
			Type:       entities.NotificationTypeTxFailed,
			APIVersion: "v1",
			Error:      errStr,
		}

		mockEventStream.EXPECT().FindOneByTenantAndChain(gomock.Any(), job.TenantID, job.ChainUUID, userInfo.AllowedTenants, userInfo.Username).Return(eventStream, nil)
		mockNotification.EXPECT().Insert(gomock.Any(), expectedNotif).Return(expectedNotif, nil)
		messenger.EXPECT().TransactionNotificationMessage(gomock.Any(), eventStream, expectedNotif, userInfo).Return(nil)

		err := usecase.Execute(ctx, job, errStr, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should do nothing if no event stream is found", func(t *testing.T) {
		job := testdata.FakeJob()

		mockEventStream.EXPECT().FindOneByTenantAndChain(gomock.Any(), job.TenantID, job.ChainUUID, userInfo.AllowedTenants, userInfo.Username).Return(nil, nil)

		err := usecase.Execute(ctx, job, errStr, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should do nothing if the job is a failed child", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Status = entities.StatusFailed
		job.InternalData.ParentJobUUID = "IHaveAParent"

		eventStream := testdata.FakeWebhookEventStream()

		mockEventStream.EXPECT().FindOneByTenantAndChain(gomock.Any(), job.TenantID, job.ChainUUID, userInfo.AllowedTenants, userInfo.Username).Return(eventStream, nil)

		err := usecase.Execute(ctx, job, errStr, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should fail with same error if cannot find stream", func(t *testing.T) {
		job := testdata.FakeJob()
		expectedErr := errors.NotFoundError("error")

		mockEventStream.EXPECT().FindOneByTenantAndChain(gomock.Any(), job.TenantID, job.ChainUUID, userInfo.AllowedTenants, userInfo.Username).Return(nil, expectedErr)

		err := usecase.Execute(ctx, job, errStr, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(notifyTransactionComponent), err)
	})
}
