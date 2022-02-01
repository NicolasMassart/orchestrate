// +build unit

package jobs

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	mock2 "github.com/consensys/orchestrate/pkg/toolkit/app/metrics/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	mocks2 "github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	"github.com/consensys/orchestrate/src/api/metrics/mock"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"

	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateJob_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockDBTX := mocks.NewMockTx(ctrl)
	mockTransactionDA := mocks.NewMockTransactionAgent(ctrl)
	mockJobDA := mocks.NewMockJobAgent(ctrl)
	mockLogDA := mocks.NewMockLogAgent(ctrl)
	mockUpdateChilrenUC := mocks2.NewMockUpdateChildrenUseCase(ctrl)
	mockStartNextJobUC := mocks2.NewMockStartNextJobUseCase(ctrl)
	mockMetrics := mock.NewMockTransactionSchedulerMetrics(ctrl)

	jobsLatencyHistogram := mock2.NewMockHistogram(ctrl)
	jobsLatencyHistogram.EXPECT().With(gomock.Any()).AnyTimes().Return(jobsLatencyHistogram)
	jobsLatencyHistogram.EXPECT().Observe(gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().JobsLatencyHistogram().AnyTimes().Return(jobsLatencyHistogram)

	minedLatencyHistogram := mock2.NewMockHistogram(ctrl)
	minedLatencyHistogram.EXPECT().With(gomock.Any()).AnyTimes().Return(minedLatencyHistogram)
	minedLatencyHistogram.EXPECT().Observe(gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().MinedLatencyHistogram().AnyTimes().Return(minedLatencyHistogram)

	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()
	mockDB.EXPECT().Begin().Return(mockDBTX, nil).AnyTimes()
	mockDB.EXPECT().Transaction().Return(mockTransactionDA).AnyTimes()
	mockDBTX.EXPECT().Job().Return(mockJobDA).AnyTimes()
	mockDBTX.EXPECT().Log().Return(mockLogDA).AnyTimes()
	mockDBTX.EXPECT().Transaction().Return(mockTransactionDA).AnyTimes()
	mockDBTX.EXPECT().Commit().Return(nil).AnyTimes()
	mockDBTX.EXPECT().Rollback().Return(nil).AnyTimes()
	mockDBTX.EXPECT().Close().Return(nil).AnyTimes()
	mockUpdateChilrenUC.EXPECT().WithDBTransaction(mockDBTX).Return(mockUpdateChilrenUC).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewUpdateJobUseCase(mockDB, mockUpdateChilrenUC, mockStartNextJobUC, mockMetrics)

	nextStatus := entities.StatusStarted
	logMessage := "message"
	ctx := context.Background()

	t.Run("should execute use case successfully", func(t *testing.T) {
		jobEntity := testdata.FakeJob()

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, true).
			Return(jobEntity, nil)
		mockTransactionDA.EXPECT().Update(gomock.Any(), jobEntity.Transaction, jobEntity.UUID).Return(nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, jobModelUpdate *entities.Job) error {
			assert.Equal(t, jobModelUpdate.InternalData, jobEntity.InternalData)
			assert.Equal(t, jobModelUpdate.Labels, jobEntity.Labels)
			return nil
		})
		mockLogDA.EXPECT().Insert(gomock.Any(), &entities.Log{
			Status:  nextStatus,
			Message: logMessage,
		}, jobEntity.UUID).Return(nil)

		_, err := usecase.Execute(ctx, jobEntity, nextStatus, logMessage, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should execute use case successfully if transaction is empty", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		jobEntity.Transaction = nil
	
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, true).
			Return(jobEntity, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, jobModelUpdate *entities.Job) error {
			assert.Equal(t, jobModelUpdate.InternalData, jobEntity.InternalData)
			assert.Equal(t, jobModelUpdate.Labels, jobEntity.Labels)
			return nil
		})
		mockLogDA.EXPECT().Insert(gomock.Any(), &entities.Log{
			Status:  nextStatus,
			Message: logMessage,
		}, jobEntity.UUID).Return(nil)
	
		_, err := usecase.Execute(ctx, jobEntity, nextStatus, logMessage, userInfo)
	
		assert.NoError(t, err)
	})
	
	t.Run("should execute use case successfully if status is empty", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
	
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, true).
			Return(jobEntity, nil)
		mockTransactionDA.EXPECT().Update(gomock.Any(), jobEntity.Transaction, jobEntity.UUID).Return(nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, jobModelUpdate *entities.Job) error {
			assert.Equal(t, jobModelUpdate.InternalData, jobEntity.InternalData)
			assert.Equal(t, jobModelUpdate.Labels, jobEntity.Labels)
			return nil
		})
	
		_, err := usecase.Execute(ctx, jobEntity, "", "", userInfo)
	
		assert.NoError(t, err)
	})
	
	t.Run("should execute use case successfully if status is PENDING", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		jobEntity.Status = entities.StatusStarted
		jobEntity.Transaction = nil
		status := entities.StatusPending
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, true).
			Return(jobEntity, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, jobModelUpdate *entities.Job) error {
			assert.Equal(t, jobModelUpdate.InternalData, jobEntity.InternalData)
			assert.Equal(t, jobModelUpdate.Labels, jobEntity.Labels)
			return nil
		})
		mockLogDA.EXPECT().Insert(gomock.Any(), &entities.Log{
			Status:  status,
			Message: logMessage,
		}, jobEntity.UUID).Return(nil)
	
		_, err := usecase.Execute(ctx, jobEntity, status, logMessage, userInfo)
		assert.NoError(t, err)
	})
	
	t.Run("should execute use case successfully if status is MINED", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		jobEntity.Transaction = nil
		jobEntity.Status = entities.StatusPending
		status := entities.StatusMined
	
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, true).
			Return(jobEntity, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, jobModelUpdate *entities.Job) error {
			assert.Equal(t, jobModelUpdate.InternalData, jobEntity.InternalData)
			assert.Equal(t, jobModelUpdate.Labels, jobEntity.Labels)
			return nil
		})
		mockLogDA.EXPECT().Insert(gomock.Any(), &entities.Log{
			Status:  status,
			Message: logMessage,
		}, jobEntity.UUID).Return(nil)
	
		_, err := usecase.Execute(ctx, jobEntity, status, logMessage, userInfo)
		assert.NoError(t, err)
	})
	
	t.Run("should execute use case successfully if status is MINED and update all the children jobs", func(t *testing.T) {
		jobParentEntity := testdata.FakeJob()
		jobEntity := testdata.FakeJob()
		jobEntity.Transaction = nil
		jobEntity.Status = entities.StatusPending
		jobEntity.InternalData.ParentJobUUID = jobParentEntity.UUID
	
		nextStatus := entities.StatusMined
	
		mockJobDA.EXPECT().LockOneByUUID(gomock.Any(), jobParentEntity.UUID).Return(nil)
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, true).
			Return(jobEntity, nil)
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(jobEntity, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, jobModelUpdate *entities.Job) error {
			assert.Equal(t, jobModelUpdate.InternalData, jobEntity.InternalData)
			assert.Equal(t, jobModelUpdate.Labels, jobEntity.Labels)
			return nil
		})
		mockLogDA.EXPECT().Insert(gomock.Any(), &entities.Log{
			Status:  nextStatus,
			Message: logMessage,
		}, jobEntity.UUID).Return(nil)
		mockUpdateChilrenUC.EXPECT().
			Execute(gomock.Any(), jobEntity.UUID, jobParentEntity.UUID, entities.StatusNeverMined, userInfo).
			Return(nil)
	
		_, err := usecase.Execute(ctx, jobEntity, nextStatus, logMessage, userInfo)
		assert.NoError(t, err)
	})
	
	t.Run("should fail with InvalidParameterError if status is MINED", func(t *testing.T) {
		jobModel := testdata.FakeJob()
		jobModel.Status = entities.StatusMined
	
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), jobModel.UUID, userInfo.AllowedTenants, userInfo.Username, true).
			Return(jobModel, nil)
	
		_, err := usecase.Execute(ctx, jobModel, nextStatus, logMessage, userInfo)
		assert.True(t, errors.IsInvalidParameterError(err))
	})
	
	t.Run("should fail with the same error if update transaction fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		jobEntity := testdata.FakeJob()
	
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username, true).
			Return(jobEntity, nil)
		mockTransactionDA.EXPECT().Update(gomock.Any(), gomock.Any(), jobEntity.UUID).Return(expectedErr)
	
		_, err := usecase.Execute(ctx, jobEntity, nextStatus, logMessage, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateJobComponent), err)
	})
	
	t.Run("should fail with the same error if update job fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		jobEntity := testdata.FakeJob()
	
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username, true).Return(jobEntity, nil)
		mockLogDA.EXPECT().Insert(gomock.Any(), gomock.Any(), jobEntity.UUID).Return(nil)
		mockTransactionDA.EXPECT().Update(gomock.Any(), gomock.Any(), jobEntity.UUID).Return(nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).Return(expectedErr)
	
		_, err := usecase.Execute(ctx, jobEntity, nextStatus, logMessage, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateJobComponent), err)
	})
	
	t.Run("should fail with InvalidStateError if status is invalid for CREATED", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		jobEntity.Transaction = nil
		jobEntity.Status = entities.StatusCreated
	
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username, true).Return(jobEntity, nil)
	
		_, err := usecase.Execute(ctx, jobEntity, entities.StatusPending, logMessage, userInfo)
		assert.True(t, errors.IsInvalidStateError(err))
	})
	
	t.Run("should fail with InvalidStateError if status is invalid for STARTED", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		jobEntity.Transaction = nil
		jobEntity.Status = entities.StatusStarted
	
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username, true).Return(jobEntity, nil)
	
		_, err := usecase.Execute(ctx, jobEntity, entities.StatusMined, logMessage, userInfo)
		assert.True(t, errors.IsInvalidStateError(err))
	})
	
	t.Run("should fail with InvalidStateError if status is invalid for PENDING", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		jobEntity.Transaction = nil
		jobEntity.Status = entities.StatusPending
	
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username, true).Return(jobEntity, nil)
	
		_, err := usecase.Execute(ctx, jobEntity, entities.StatusStarted, logMessage, userInfo)
		assert.True(t, errors.IsInvalidStateError(err))
	})
	
	t.Run("should fail with InvalidStateError if status is invalid for FAILED", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		jobEntity.Transaction = nil
		jobEntity.Status = entities.StatusCreated
	
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username, true).Return(jobEntity, nil)
		_, err := usecase.Execute(ctx, jobEntity, entities.StatusFailed, logMessage, userInfo)
		assert.True(t, errors.IsInvalidStateError(err))
	})
	
	t.Run("should allow transition of job from RESENDING to RESENDING", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		jobEntity.Transaction = nil
		jobEntity.Status = entities.StatusResending
	
		status := entities.StatusResending
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username, true).Return(jobEntity, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, jobModelUpdate *entities.Job) error {
			assert.Equal(t, jobModelUpdate.InternalData, jobEntity.InternalData)
			assert.Equal(t, jobModelUpdate.Labels, jobEntity.Labels)
			return nil
		})
		mockLogDA.EXPECT().Insert(gomock.Any(), &entities.Log{
			Status:  status,
			Message: logMessage,
		}, jobEntity.UUID).Return(nil)
		_, err := usecase.Execute(ctx, jobEntity, status, logMessage, userInfo)
		assert.NoError(t, err)
	})
	
	t.Run("should fail with the same error if insert log fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		jobEntity := testdata.FakeJob()
		jobEntity.Transaction = nil
	
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username, true).Return(jobEntity, nil)
		mockLogDA.EXPECT().Insert(gomock.Any(), gomock.Any(), jobEntity.UUID).Return(expectedErr)
	
		_, err := usecase.Execute(ctx, jobEntity, nextStatus, logMessage, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateJobComponent), err)
	})
	
	t.Run("should trigger next job start if nextStatus is STORED", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		jobEntity.Status = entities.StatusStarted
		
		nextJobEntity := testdata.FakeJob()
		jobEntity.NextJobUUID = nextJobEntity.UUID
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, true).Return(jobEntity, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		mockTransactionDA.EXPECT().Update(gomock.Any(), gomock.Any(), jobEntity.UUID).Return(nil)
		mockLogDA.EXPECT().Insert(gomock.Any(), gomock.Any(), jobEntity.UUID).Return(nil)
		mockStartNextJobUC.EXPECT().Execute(gomock.Any(), jobEntity.UUID, userInfo).Return(nil)
	
		_, err := usecase.Execute(ctx, jobEntity, entities.StatusStored, "", userInfo)
	
		assert.NoError(t, err)
	})
}
