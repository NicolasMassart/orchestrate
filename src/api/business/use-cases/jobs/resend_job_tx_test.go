// +build unit

package jobs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mocks3 "github.com/consensys/orchestrate/src/infra/messenger/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestResendJobTx_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobDA := mocks.NewMockJobAgent(ctrl)
	kafkaProducer := mocks3.NewMockProducer(ctrl)
	db := mocks.NewMockDB(ctrl)
	db.EXPECT().Job().Return(jobDA).AnyTimes()

	topicSender := "topic-sender"

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewResendJobTxUseCase(db, kafkaProducer, topicSender)

	t.Run("should execute use case successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Status = entities.StatusPending
		job.Logs = append(job.Logs, &entities.Log{
			Status:    entities.StatusPending,
			CreatedAt: time.Now().Add(time.Second),
		})

		jobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(job, nil)

		kafkaProducer.EXPECT().SendJobMessage(topicSender, job, userInfo).Return(nil)

		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.NoError(t, err)
	})

	t.Run("should fail with same error if FindOne fails", func(t *testing.T) {
		job := testdata.FakeJob()
		expectedErr := errors.NotFoundError("error")

		jobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(nil, expectedErr)

		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(resendJobTxComponent), err)
	})

	t.Run("should fail with KafkaConnectionError if Produce fails", func(t *testing.T) {
		expectedErr := fmt.Errorf("error")
		job := testdata.FakeJob()
		job.Status = entities.StatusPending
		job.Logs = append(job.Logs, &entities.Log{
			Status:    entities.StatusPending,
			CreatedAt: time.Now().Add(time.Second),
		})

		jobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).Return(job, nil)
		kafkaProducer.EXPECT().SendJobMessage(topicSender, job, userInfo).Return(expectedErr)
		err := usecase.Execute(ctx, job.UUID, userInfo)
		assert.Equal(t, expectedErr, err)
	})
}
