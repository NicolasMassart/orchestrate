// +build unit

package jobs

import (
	"context"
	"testing"

	mock3 "github.com/consensys/orchestrate/pkg/sdk/mock"
	mock2 "github.com/consensys/orchestrate/pkg/toolkit/app/metrics/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	mocks2 "github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	"github.com/consensys/orchestrate/src/api/metrics/mock"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateJob_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	jobDA := mocks.NewMockJobAgent(ctrl)
	chainDA := mocks.NewMockChainAgent(ctrl)
	startNextJobUC := mocks2.NewMockStartNextJobUseCase(ctrl)
	metrics := mock.NewMockTransactionSchedulerMetrics(ctrl)
	notifyTxUC := mocks2.NewMockNotifyTransactionUseCase(ctrl)
	
	messengerTxListener := mock3.NewMockMessengerTxListener(ctrl)

	jobsLatencyHistogram := mock2.NewMockHistogram(ctrl)
	jobsLatencyHistogram.EXPECT().With(gomock.Any()).AnyTimes().Return(jobsLatencyHistogram)
	jobsLatencyHistogram.EXPECT().Observe(gomock.Any()).AnyTimes()
	metrics.EXPECT().JobsLatencyHistogram().AnyTimes().Return(jobsLatencyHistogram)

	minedLatencyHistogram := mock2.NewMockHistogram(ctrl)
	minedLatencyHistogram.EXPECT().With(gomock.Any()).AnyTimes().Return(minedLatencyHistogram)
	minedLatencyHistogram.EXPECT().Observe(gomock.Any()).AnyTimes()
	metrics.EXPECT().MinedLatencyHistogram().AnyTimes().Return(minedLatencyHistogram)

	mockDB.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, persistFunc func(dbtx store.DB) error) error {
			return persistFunc(mockDB)
		}).AnyTimes()
	mockDB.EXPECT().Job().Return(jobDA).AnyTimes()
	mockDB.EXPECT().Chain().Return(chainDA).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewUpdateJobUseCase(mockDB, startNextJobUC, metrics, notifyTxUC, messengerTxListener)

	ctx := context.Background()

	t.Run("should execute use case for PENDING status successfully", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		nextStatus := entities.StatusPending
		statusMsg := "tx pending"
		jobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, gomock.Any()).
			Times(2).Return(jobEntity, nil)

		messengerTxListener.EXPECT().PendingJobMessage(gomock.Any(), jobEntity, userInfo)
		jobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, job *entities.Job, log *entities.Log) error {
				assert.Equal(t, jobEntity.UUID, job.UUID)
				assert.Equal(t, nextStatus, job.Status)
				assert.Equal(t, nextStatus, log.Status)
				assert.Equal(t, statusMsg, log.Message)
				return nil
			})

		_, err := usecase.Execute(ctx, jobEntity, nextStatus, statusMsg, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should execute use case for MINED status successfully", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		nextJobEntity := testdata.FakeJob()
		jobEntity.NextJobUUID = nextJobEntity.UUID

		siblingJob := testdata.FakeJob()
		siblingJob.Status = entities.StatusPending
		siblingJobEntities := []*entities.Job{siblingJob}
		nextStatus := entities.StatusMined
		statusMsg := "tx mined"

		jobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, gomock.Any()).
			Times(2).Return(jobEntity, nil)

		jobDA.EXPECT().GetSiblingJobs(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username).
			Return(siblingJobEntities, nil)

		jobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, job *entities.Job, log *entities.Log) error {
				assert.Equal(t, jobEntity.UUID, job.UUID)
				assert.Equal(t, nextStatus, job.Status)
				assert.Equal(t, nextStatus, log.Status)
				assert.Equal(t, statusMsg, log.Message)
				return nil
			})

		jobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, job *entities.Job, log *entities.Log) error {
				assert.Equal(t, siblingJobEntities[0].UUID, job.UUID)
				assert.Equal(t, entities.StatusNeverMined, job.Status)
				assert.Equal(t, entities.StatusNeverMined, log.Status)
				return nil
			})

		notifyTxUC.EXPECT().Execute(gomock.Any(), jobEntity, "", userInfo).Return(nil)

		startNextJobUC.EXPECT().Execute(gomock.Any(), jobEntity.UUID, userInfo).Return(nil)

		_, err := usecase.Execute(ctx, jobEntity, nextStatus, statusMsg, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should execute use case for FAILED status successfully", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		nextStatus := entities.StatusFailed
		statusMsg := "tx failed"
		jobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, gomock.Any()).
			Times(2).Return(jobEntity, nil)

		jobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, job *entities.Job, log *entities.Log) error {
				assert.Equal(t, jobEntity.UUID, job.UUID)
				assert.Equal(t, nextStatus, job.Status)
				assert.Equal(t, nextStatus, log.Status)
				assert.Equal(t, statusMsg, log.Message)
				return nil
			})
		notifyTxUC.EXPECT().Execute(gomock.Any(), jobEntity, statusMsg, userInfo).Return(nil)

		_, err := usecase.Execute(ctx, jobEntity, nextStatus, statusMsg, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should execute use case for STORED status successfully", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		nextStatus := entities.StatusStored
		statusMsg := "tx stored"
		nextJobEntity := testdata.FakeJob()
		jobEntity.NextJobUUID = nextJobEntity.UUID

		jobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, gomock.Any()).
			Times(2).Return(jobEntity, nil)

		jobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, job *entities.Job, log *entities.Log) error {
				assert.Equal(t, jobEntity.UUID, job.UUID)
				assert.Equal(t, nextStatus, job.Status)
				assert.Equal(t, nextStatus, log.Status)
				assert.Equal(t, statusMsg, log.Message)
				return nil
			})

		startNextJobUC.EXPECT().Execute(gomock.Any(), jobEntity.UUID, userInfo).Return(nil)
		_, err := usecase.Execute(ctx, jobEntity, nextStatus, statusMsg, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should execute use case for RESENDING status successfully", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		nextStatus := entities.StatusResending

		jobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, gomock.Any()).
			Times(2).Return(jobEntity, nil)

		jobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, job *entities.Job, log *entities.Log) error {
				assert.Equal(t, jobEntity.UUID, job.UUID)
				assert.Equal(t, jobEntity.Status, job.Status)
				assert.Equal(t, nextStatus, log.Status)
				return nil
			})

		_, err := usecase.Execute(ctx, jobEntity, nextStatus, "", userInfo)

		assert.NoError(t, err)
	})

	t.Run("should execute use case for WARNING status successfully", func(t *testing.T) {
		jobEntity := testdata.FakeJob()
		nextStatus := entities.StatusWarning

		jobDA.EXPECT().FindOneByUUID(gomock.Any(), jobEntity.UUID, userInfo.AllowedTenants, userInfo.Username, gomock.Any()).
			Times(2).Return(jobEntity, nil)

		jobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, job *entities.Job, log *entities.Log) error {
				assert.Equal(t, jobEntity.UUID, job.UUID)
				assert.Equal(t, jobEntity.Status, job.Status)
				assert.Equal(t, nextStatus, log.Status)
				return nil
			})

		_, err := usecase.Execute(ctx, jobEntity, nextStatus, "", userInfo)

		assert.NoError(t, err)
	})
}
