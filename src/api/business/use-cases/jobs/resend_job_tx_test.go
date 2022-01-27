// +build unit

package jobs

import (
	"context"
	"fmt"
	"testing"
	"time"

	mocks2 "github.com/Shopify/sarama/mocks"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/consensys/orchestrate/src/infra/broker/sarama"
	encoding "github.com/consensys/orchestrate/pkg/encoding/proto"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/pkg/types/tx"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/api/store/models/testdata"
)

func TestResendJobTx_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJobDA := mocks.NewMockJobAgent(ctrl)
	mockKafkaProducer := mocks2.NewSyncProducer(t, nil)
	mockDB := mocks.NewMockDB(ctrl)
	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewResendJobTxUseCase(mockDB, mockKafkaProducer, sarama.NewKafkaTopicConfig(viper.GetViper()))

	t.Run("should execute use case successfully", func(t *testing.T) {
		job := testdata.FakeJobModel(1)
		job.ID = 1
		job.UUID = "6380e2b6-b828-43ee-abdc-de0f8d57dc5f"
		job.Transaction.Sender = "0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18"
		job.Schedule = testdata.FakeSchedule("", "")
		job.Status = entities.StatusPending
		job.Logs = append(job.Logs, &models.Log{
			ID:        1,
			Status:    entities.StatusPending,
			CreatedAt: time.Now().Add(time.Second),
		})

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
			assert.Equal(t, envelope.GetParentJobUUID(), job.UUID)
			return nil
		})

		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.NoError(t, err)
	})

	t.Run("should fail with same error if FindOne fails", func(t *testing.T) {
		job := testdata.FakeJobModel(1)
		job.UUID = "6380e2b6-b828-43ee-abdc-de0f8d57dc5f"
		expectedErr := errors.NotFoundError("error")

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(nil, expectedErr)

		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(resendJobTxComponent), err)
	})

	t.Run("should fail with KafkaConnectionError if Produce fails", func(t *testing.T) {
		job := testdata.FakeJobModel(1)
		job.UUID = "6380e2b6-b828-43ee-abdc-de0f8d57dc5f"
		job.Transaction.Sender = "0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18"
		job.Schedule = testdata.FakeSchedule("", "")
		job.Status = entities.StatusPending
		job.Logs = append(job.Logs, &models.Log{
			ID:        1,
			Status:    entities.StatusPending,
			CreatedAt: time.Now().Add(time.Second),
		})

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(job, nil)
		mockKafkaProducer.ExpectSendMessageAndFail(fmt.Errorf("error"))
		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.True(t, errors.IsKafkaConnectionError(err))
	})
}
