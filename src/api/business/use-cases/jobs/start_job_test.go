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

	mocks2 "github.com/Shopify/sarama/mocks"
	encoding "github.com/consensys/orchestrate/pkg/encoding/proto"
	"github.com/consensys/orchestrate/pkg/types/tx"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/orchestrate/src/infra/broker/sarama"
	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestStartJob_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJobDA := mocks.NewMockJobAgent(ctrl)
	mockLogDA := mocks.NewMockLogAgent(ctrl)
	mockKafkaProducer := mocks2.NewSyncProducer(t, nil)
	mockMetrics := mock.NewMockTransactionSchedulerMetrics(ctrl)

	jobsLatencyHistogram := mock2.NewMockHistogram(ctrl)
	jobsLatencyHistogram.EXPECT().With(gomock.Any()).AnyTimes().Return(jobsLatencyHistogram)
	jobsLatencyHistogram.EXPECT().Observe(gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().JobsLatencyHistogram().AnyTimes().Return(jobsLatencyHistogram)

	mockDB := mocks.NewMockDB(ctrl)

	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()
	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()
	mockDB.EXPECT().Log().Return(mockLogDA).AnyTimes()
	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewStartJobUseCase(mockDB, mockKafkaProducer, sarama.NewKafkaTopicConfig(viper.GetViper()), mockMetrics)

	t.Run("should execute use case successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(job, nil)
		mockKafkaProducer.ExpectSendMessageWithCheckerFunctionAndSucceed(func(val []byte) error {
			txEnvelope := &tx.TxEnvelope{}
			err := encoding.Unmarshal(val, txEnvelope)
			if err != nil {
				return err
			}
			envelope, err := txEnvelope.Envelope()
			if err != nil {
				return err
			}

			assert.Equal(t, envelope.GetJobUUID(), job.UUID)
			assert.False(t, envelope.IsOneTimeKeySignature())
			return nil
		})
		mockDB.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(nil)

		err := usecase.Execute(ctx, job.UUID, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should execute use case with one-time-key successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.InternalData = &entities.InternalData{
			OneTimeKey: true,
		}

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(job, nil)
		mockKafkaProducer.ExpectSendMessageWithCheckerFunctionAndSucceed(func(val []byte) error {
			txEnvelope := &tx.TxEnvelope{}
			err := encoding.Unmarshal(val, txEnvelope)
			if err != nil {
				return err
			}
			envelope, err := txEnvelope.Envelope()
			if err != nil {
				return err
			}

			assert.Equal(t, envelope.GetJobUUID(), job.UUID)
			assert.True(t, envelope.IsOneTimeKeySignature())
			return nil
		})
		mockDB.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(nil)

		err := usecase.Execute(ctx, job.UUID, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should fail with same error if FindOne fails", func(t *testing.T) {
		job := testdata.FakeJob()
		expectedErr := errors.NotFoundError("error")

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(nil, expectedErr)

		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(startJobComponent), err)
	})

	t.Run("should fail with same error if RunInTransaction fails", func(t *testing.T) {
		job := testdata.FakeJob()
		expectedErr := errors.PostgresConnectionError("error")

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(job, nil)
		mockDB.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(expectedErr)

		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(startJobComponent), err)
	})

	t.Run("should fail with KafkaConnectionError if Produce fails", func(t *testing.T) {
		job := testdata.FakeJob()

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(job, nil)
		mockDB.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(nil)
		mockKafkaProducer.ExpectSendMessageAndFail(fmt.Errorf("error"))
		mockDB.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(nil)

		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.True(t, errors.IsKafkaConnectionError(err))
	})
}
