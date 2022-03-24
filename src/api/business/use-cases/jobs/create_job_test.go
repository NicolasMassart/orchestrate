// +build unit

package jobs

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	mocks2 "github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateJob_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockScheduleDA := mocks.NewMockScheduleAgent(ctrl)
	mockAccountDA := mocks.NewMockAccountAgent(ctrl)
	mockJobDA := mocks.NewMockJobAgent(ctrl)
	mockGetChainUC := mocks2.NewMockGetChainUseCase(ctrl)

	mockDB.EXPECT().Schedule().Return(mockScheduleDA).AnyTimes()
	mockDB.EXPECT().Account().Return(mockAccountDA).AnyTimes()
	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()

	qkmStoreID := "qkm-store-id"
	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewCreateJobUseCase(mockDB, mockGetChainUC, qkmStoreID)
	fakeChain := testdata.FakeChain()
	fakeAccount := testdata.FakeAccount()

	fakeSchedule := testdata.FakeSchedule()

	t.Run("should execute use case successfully", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		jobEntity.ScheduleUUID = fakeSchedule.UUID

		mockGetChainUC.EXPECT().Execute(gomock.Any(), jobEntity.ChainUUID, userInfo).Return(fakeChain, nil)
		mockAccountDA.EXPECT().FindOneByAddress(gomock.Any(), jobEntity.Transaction.From.String(),
			userInfo.AllowedTenants, userInfo.Username).
			Return(fakeAccount, nil)
		mockScheduleDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.ScheduleUUID, userInfo.AllowedTenants, userInfo.Username).
			Return(fakeSchedule, nil)
		mockJobDA.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, job *entities.Job, log *entities.Log) error {
				assert.Equal(t, job.TenantID, userInfo.TenantID)
				assert.Equal(t, job.OwnerID, userInfo.Username)
				assert.Equal(t, job.InternalData.ChainID, fakeChain.ChainID)
				assert.Equal(t, log.Status, entities.StatusCreated)
				return nil
			})

		_, err := usecase.Execute(context.Background(), jobEntity, userInfo)

		assert.NoError(t, err)
	})
	
	t.Run("should fail with InvalidParameterError if chain is not found", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
	
		mockGetChainUC.EXPECT().Execute(gomock.Any(), jobEntity.ChainUUID, userInfo).
			Return(nil, errors.NotFoundError("error"))
	
		_, err := usecase.Execute(context.Background(), jobEntity, userInfo)
		assert.True(t, errors.IsInvalidParameterError(err))
	})
	
	t.Run("should fail with InvalidParameterError if account does not exist", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
	
		mockGetChainUC.EXPECT().Execute(gomock.Any(), jobEntity.ChainUUID, userInfo).Return(fakeChain, nil)
		mockAccountDA.EXPECT().FindOneByAddress(gomock.Any(), jobEntity.Transaction.From.String(), userInfo.AllowedTenants, userInfo.Username).
			Return(nil, errors.NotFoundError("error"))
	
		_, err := usecase.Execute(context.Background(), jobEntity, userInfo)
		assert.True(t, errors.IsInvalidParameterError(err))
	})
	
	t.Run("should fail with InvalidParameterError if schedule is not found", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		jobEntity.ScheduleUUID = fakeSchedule.UUID
	
		mockGetChainUC.EXPECT().Execute(gomock.Any(), jobEntity.ChainUUID, userInfo).Return(fakeChain, nil)
		mockAccountDA.EXPECT().FindOneByAddress(gomock.Any(), jobEntity.Transaction.From.String(), userInfo.AllowedTenants, userInfo.Username).
			Return(fakeAccount, nil)
		mockScheduleDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.ScheduleUUID, userInfo.AllowedTenants, userInfo.Username).Return(nil, errors.NotFoundError("error"))
	
		_, err := usecase.Execute(context.Background(), jobEntity, userInfo)
		assert.True(t, errors.IsInvalidParameterError(err))
	})
	

	t.Run("should fail with same error if job fails to insert", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		jobEntity := testdata.FakeJob()
		jobEntity.ScheduleUUID = fakeSchedule.UUID
	
		mockGetChainUC.EXPECT().Execute(gomock.Any(), jobEntity.ChainUUID, userInfo).Return(fakeChain, nil)
		mockAccountDA.EXPECT().FindOneByAddress(gomock.Any(), jobEntity.Transaction.From.String(), userInfo.AllowedTenants, userInfo.Username).
			Return(fakeAccount, nil)
		mockScheduleDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.ScheduleUUID, userInfo.AllowedTenants, userInfo.Username).
			Return(fakeSchedule, nil)
		mockJobDA.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr)
	
		_, err := usecase.Execute(context.Background(), jobEntity, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(createJobComponent), err)
	})
}
