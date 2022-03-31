// +build unit

package jobs

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	mock2 "github.com/consensys/orchestrate/pkg/toolkit/app/metrics/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/metrics/mock"
	"github.com/consensys/orchestrate/src/entities"
	mocks3 "github.com/consensys/orchestrate/src/infra/kafka/mocks"

	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestStartJob_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobDA := mocks.NewMockJobAgent(ctrl)
	kafkaProducer := mocks3.NewMockProducer(ctrl)
	metrics := mock.NewMockTransactionSchedulerMetrics(ctrl)

	jobsLatencyHistogram := mock2.NewMockHistogram(ctrl)
	jobsLatencyHistogram.EXPECT().With(gomock.Any()).AnyTimes().Return(jobsLatencyHistogram)
	jobsLatencyHistogram.EXPECT().Observe(gomock.Any()).AnyTimes()
	metrics.EXPECT().JobsLatencyHistogram().AnyTimes().Return(jobsLatencyHistogram)

	db := mocks.NewMockDB(ctrl)
	db.EXPECT().Job().Return(jobDA).AnyTimes()

	topicsCfg := sarama.NewDefaultTopicConfig()
	
	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewStartJobUseCase(db, kafkaProducer, topicsCfg, metrics)

	t.Run("should execute use case successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		jobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(job, nil)
		kafkaProducer.EXPECT().SendJobMessage(topicsCfg.Sender, job, userInfo).Return(nil)
		jobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, job *entities.Job, log *entities.Log) error {
				assert.Equal(t, job.Status, entities.StatusStarted)
				assert.Equal(t, log.Status, entities.StatusStarted)
				return nil
			})

		err := usecase.Execute(ctx, job.UUID, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should execute use case with one-time-key successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.InternalData = &entities.InternalData{
			OneTimeKey: true,
		}

		jobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(job, nil)
		kafkaProducer.EXPECT().SendJobMessage(topicsCfg.Sender, job, userInfo).Return(nil)
		jobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, job *entities.Job, log *entities.Log) error {
				assert.Equal(t, job.Status, entities.StatusStarted)
				assert.Equal(t, log.Status, entities.StatusStarted)
				return nil
			})

		err := usecase.Execute(ctx, job.UUID, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should fail with same error if FindOne fails", func(t *testing.T) {
		job := testdata.FakeJob()
		expectedErr := errors.NotFoundError("error")

		jobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(nil, expectedErr)

		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(startJobComponent), err)
	})

	t.Run("should fail with same error if RunInTransaction fails", func(t *testing.T) {
		job := testdata.FakeJob()
		expectedErr := errors.PostgresConnectionError("error")

		jobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(job, nil)
		jobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(startJobComponent), err)
	})

	t.Run("should fail with KafkaConnectionError if Produce fails", func(t *testing.T) {
		expectedErr := fmt.Errorf("error")
		job := testdata.FakeJob()

		jobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(job, nil)
		jobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		kafkaProducer.EXPECT().SendJobMessage(topicsCfg.Sender, job, userInfo).Return(expectedErr)

		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.Equal(t, expectedErr, err)
	})
}
