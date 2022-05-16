// +build unit

package txsentry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mocks2 "github.com/consensys/orchestrate/src/tx-listener/store/mocks"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var errMsgExceedTime = "exceeded waiting time for stopping"
var extendedWaitingTime = time.Millisecond * 500
var defaultRetryInterval = time.Second

func TestRetryJobSession_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgAPI := mock.NewMockMessengerAPI(ctrl)
	jobAPIClient := mock.NewMockJobClient(ctrl)
	retryJobUseCase := mocks.NewMockRetryJob(ctrl)
	pendingJobState := mocks2.NewMockPendingJob(ctrl)
	logger := log.NewLogger()

	t.Run("should start and exit gracefully after running retry session twice", func(t *testing.T) {
		job := testdata.FakeJob()
		childJob := testdata.FakeJob()
		job.InternalData.RetryInterval = defaultRetryInterval
		cStopErr := make(chan error, 1)

		jobAPIClient.EXPECT().SearchJob(gomock.Any(), gomock.Any()).Return([]*api.JobResponse{}, nil)
		pendingJobState.EXPECT().GetByTxHash(gomock.Any(), job.ChainUUID, job.Transaction.Hash).
			Times(2).Return(job, nil)
		prevRetryCall := retryJobUseCase.EXPECT().Execute(gomock.Any(), job, job.UUID, 0).Return(childJob.UUID, nil)
		retryJobUseCase.EXPECT().Execute(gomock.Any(), job, childJob.UUID, 1).After(prevRetryCall).
			Return("", nil)

		usecase := NewRetryJobSession(msgAPI, jobAPIClient, retryJobUseCase, pendingJobState, job, logger)
		go func() {
			err := usecase.Start(ctx)
			cStopErr <- err
		}()

		time.Sleep((job.InternalData.RetryInterval * 2) + extendedWaitingTime)
		
		usecase.Stop()
		
		select {
		case <-time.Tick(extendedWaitingTime):
			t.Error(errMsgExceedTime)
		case err := <-cStopErr:
			assert.NoError(t, err)
		}
	})
	
	t.Run("should start and exit gracefully after there is not pending job session twice", func(t *testing.T) {
		job := testdata.FakeJob()
		childJob := testdata.FakeJob()
		job.InternalData.RetryInterval = defaultRetryInterval
		cStopErr := make(chan error, 1)

		jobAPIClient.EXPECT().SearchJob(gomock.Any(), gomock.Any()).Return([]*api.JobResponse{}, nil)
		prevPendingCall := pendingJobState.EXPECT().GetByTxHash(gomock.Any(), job.ChainUUID, job.Transaction.Hash).Return(job, nil)
		pendingJobState.EXPECT().GetByTxHash(gomock.Any(), job.ChainUUID, job.Transaction.Hash).After(prevPendingCall).Return(nil, errors.NotFoundError(""))
		retryJobUseCase.EXPECT().Execute(gomock.Any(), job, job.UUID, 0).Return(childJob.UUID, nil)

		usecase := NewRetryJobSession(msgAPI, jobAPIClient, retryJobUseCase, pendingJobState, job, logger)
		go func() {
			err := usecase.Start(ctx)
			cStopErr <- err
		}()

		time.Sleep((job.InternalData.RetryInterval * 2) + extendedWaitingTime)
		
		select {
		case <-time.Tick(extendedWaitingTime):
			t.Error(errMsgExceedTime)
		case err := <-cStopErr:
			assert.NoError(t, err)
		}
	})

	t.Run("should start and exit with errors after running retry session fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.InternalData.RetryInterval = defaultRetryInterval
		cStopErr := make(chan error, 1)

		expectedErr := fmt.Errorf("failed to retry")
		jobAPIClient.EXPECT().SearchJob(gomock.Any(), gomock.Any()).Return([]*api.JobResponse{}, nil)
		pendingJobState.EXPECT().GetByTxHash(gomock.Any(), job.ChainUUID, job.Transaction.Hash).Return(job, nil)
		retryJobUseCase.EXPECT().Execute(gomock.Any(), job, job.UUID, 0).Return("", expectedErr)

		usecase := NewRetryJobSession(msgAPI, jobAPIClient, retryJobUseCase, pendingJobState, job, logger)
		go func() {
			err := usecase.Start(ctx)
			cStopErr <- err
		}()

		time.Sleep(job.InternalData.RetryInterval + extendedWaitingTime)
		select {
		case <-time.Tick(extendedWaitingTime):
			t.Error(errMsgExceedTime)
		case err := <-cStopErr:
			assert.Error(t, err)
		}
	})
}
