// +build unit

package jobs

import (
	"context"
	"fmt"
	"testing"

	mock2 "github.com/consensys/orchestrate/pkg/toolkit/app/metrics/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/api/metrics/mock"

	"github.com/consensys/orchestrate/src/infra/broker/sarama"
	encoding "github.com/consensys/orchestrate/pkg/encoding/proto"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/types/tx"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/api/store/models/testdata"
	mocks2 "github.com/Shopify/sarama/mocks"
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
	mockDBTX := mocks.NewMockTx(ctrl)
	mockKafkaProducer := mocks2.NewSyncProducer(t, nil)
	mockMetrics := mock.NewMockTransactionSchedulerMetrics(ctrl)

	jobsLatencyHistogram := mock2.NewMockHistogram(ctrl)
	jobsLatencyHistogram.EXPECT().With(gomock.Any()).AnyTimes().Return(jobsLatencyHistogram)
	jobsLatencyHistogram.EXPECT().Observe(gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().JobsLatencyHistogram().AnyTimes().Return(jobsLatencyHistogram)

	mockDB := mocks.NewMockDB(ctrl)
	mockDB.EXPECT().Begin().Return(mockDBTX, nil).AnyTimes()
	mockDBTX.EXPECT().Close().Return(nil).AnyTimes()

	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()
	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()
	mockDBTX.EXPECT().Log().Return(mockLogDA).AnyTimes()
	mockDBTX.EXPECT().Job().Return(mockJobDA).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewStartJobUseCase(mockDB, mockKafkaProducer, sarama.NewKafkaTopicConfig(viper.GetViper()), mockMetrics)

	t.Run("should execute use case successfully", func(t *testing.T) {
		job := testdata.FakeJobModel(1)
		job.ID = 1
		job.UUID = "6380e2b6-b828-43ee-abdc-de0f8d57dc5f"
		job.Transaction.Sender = "0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18"
		job.Schedule = testdata.FakeSchedule("", "")

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

		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		mockLogDA.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil)
		mockDBTX.EXPECT().Commit().Return(nil)
		err := usecase.Execute(ctx, job.UUID, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should execute use case with one-time-key successfully", func(t *testing.T) {
		job := testdata.FakeJobModel(1)
		job.ID = 1
		job.UUID = "6380e2b6-b828-43ee-abdc-de0f8d57dc5f"
		job.Transaction.Sender = "0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18"
		job.Schedule = testdata.FakeSchedule("", "")
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

		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		mockLogDA.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil)
		mockDBTX.EXPECT().Commit().Return(nil)
		err := usecase.Execute(ctx, job.UUID, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should fail with same error if FindOne fails", func(t *testing.T) {
		job := testdata.FakeJobModel(1)
		job.UUID = "6380e2b6-b828-43ee-abdc-de0f8d57dc5f"
		expectedErr := errors.NotFoundError("error")

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(nil, expectedErr)

		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(startJobComponent), err)
	})

	t.Run("should fail with same error if Insert log fails", func(t *testing.T) {
		job := testdata.FakeJobModel(1)
		job.ID = 1
		job.UUID = "6380e2b6-b828-43ee-abdc-de0f8d57dc5f"
		job.Transaction.Sender = "0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18"
		job.Transaction.ID = 1
		job.Schedule = testdata.FakeSchedule("", "")
		job.Schedule.ID = 1
		expectedErr := errors.PostgresConnectionError("error")

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(job, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		mockLogDA.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(expectedErr)
		mockDBTX.EXPECT().Rollback().Return(nil)
		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(startJobComponent), err)
	})

	t.Run("should fail with KafkaConnectionError if Produce fails", func(t *testing.T) {
		job := testdata.FakeJobModel(1)
		job.UUID = "6380e2b6-b828-43ee-abdc-de0f8d57dc5f"
		job.Transaction.Sender = "0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18"
		job.Schedule = testdata.FakeSchedule("", "")

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(job, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(2)
		mockLogDA.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil).Times(2)
		mockDBTX.EXPECT().Commit().Return(nil).Times(2)
		mockKafkaProducer.ExpectSendMessageAndFail(fmt.Errorf("error"))
		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.True(t, errors.IsKafkaConnectionError(err))
	})
}
