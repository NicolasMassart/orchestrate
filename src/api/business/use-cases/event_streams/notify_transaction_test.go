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

	mockDB := mocks.NewMockEventStreamAgent(ctrl)
	messenger := mock.NewMockMessengerNotifier(ctrl)
	searchContractsUC := mocks3.NewMockSearchContractUseCase(ctrl)
	decodeLogUC := mocks3.NewMockDecodeEventLogUseCase(ctrl)

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	errStr := "error"

	usecase := NewNotifyTransactionUseCase(mockDB, searchContractsUC, decodeLogUC, messenger)

	t.Run("should execute use case successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		eventStream := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().FindOneByTenantAndChain(gomock.Any(), job.TenantID, job.ChainUUID, userInfo.AllowedTenants, userInfo.Username).Return(eventStream, nil)
		messenger.EXPECT().TransactionNotificationMessage(gomock.Any(), eventStream, job, errStr, userInfo).Return(nil)

		err := usecase.Execute(ctx, job, errStr, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should do nothing if no event stream is found", func(t *testing.T) {
		job := testdata.FakeJob()

		mockDB.EXPECT().FindOneByTenantAndChain(gomock.Any(), job.TenantID, job.ChainUUID, userInfo.AllowedTenants, userInfo.Username).Return(nil, nil)

		err := usecase.Execute(ctx, job, errStr, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should do nothing if the job is a failed child", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Status = entities.StatusFailed
		job.InternalData.ParentJobUUID = "IHaveAParent"

		eventStream := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().FindOneByTenantAndChain(gomock.Any(), job.TenantID, job.ChainUUID, userInfo.AllowedTenants, userInfo.Username).Return(eventStream, nil)

		err := usecase.Execute(ctx, job, errStr, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should fail with same error if cannot find stream", func(t *testing.T) {
		job := testdata.FakeJob()
		expectedErr := errors.NotFoundError("error")

		mockDB.EXPECT().FindOneByTenantAndChain(gomock.Any(), job.TenantID, job.ChainUUID, userInfo.AllowedTenants, userInfo.Username).Return(nil, expectedErr)

		err := usecase.Execute(ctx, job, errStr, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(notifyTransactionComponent), err)
	})
}
