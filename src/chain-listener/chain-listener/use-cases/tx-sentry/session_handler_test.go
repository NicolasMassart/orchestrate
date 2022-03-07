// +build unit
// +build !race

package txsentry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/client/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases/mocks"
	storemocks "github.com/consensys/orchestrate/src/chain-listener/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPendingJobs_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apiClient := mock.NewMockOrchestrateClient(ctrl)
	sendRetryJobUseCase := mocks.NewMockSendRetryJob(ctrl)
	retrySessionState := storemocks.NewMockRetrySessions(ctrl)
	logger := log.NewLogger()

	expectedErr := fmt.Errorf("expected_err")

	usecase := RetrySessionManager(apiClient, sendRetryJobUseCase, retrySessionState, logger)

	t.Run("should start and exit gracefully after running retry session successfully", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		childJob := testdata.FakeJob()
		parentJob.InternalData.RetryInterval = time.Second

		retrySessionState.EXPECT().Add(gomock.Any(), parentJob.UUID, parentJob)
		apiClient.EXPECT().SearchJob(gomock.Any(), gomock.Any()).Return([]*api.JobResponse{}, nil)
		sendRetryJobUseCase.EXPECT().Execute(gomock.Any(), parentJob, parentJob.UUID, 0).Return(childJob.UUID, nil)
		sendRetryJobUseCase.EXPECT().Execute(gomock.Any(), parentJob, childJob.UUID, 1).Return("", nil)
		retrySessionState.EXPECT().Remove(gomock.Any(), parentJob.UUID)
		
		err := usecase.StartSession(ctx, parentJob)
		require.NoError(t, err)
		time.Sleep(parentJob.InternalData.RetryInterval * 2)
		err = usecase.StopSession(ctx, parentJob.UUID)
		assert.NoError(t, err)
	})
	
	t.Run("should start and exit gracefully after running retry session successfully", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		childJob := testdata.FakeJob()
		parentJob.InternalData.RetryInterval = time.Second

		retrySessionState.EXPECT().Add(gomock.Any(), parentJob.UUID, parentJob)
		apiClient.EXPECT().SearchJob(gomock.Any(), gomock.Any()).Return([]*api.JobResponse{}, nil)
		sendRetryJobUseCase.EXPECT().Execute(gomock.Any(), parentJob, parentJob.UUID, 0).Return(childJob.UUID, nil)
		sendRetryJobUseCase.EXPECT().Execute(gomock.Any(), parentJob, childJob.UUID, 1).Return("", nil)
		retrySessionState.EXPECT().ListByChainUUID(gomock.Any(), parentJob.ChainUUID).Return([]string{parentJob.UUID}, nil)
		retrySessionState.EXPECT().Remove(gomock.Any(), parentJob.UUID)
		
		err := usecase.StartSession(ctx, parentJob)
		require.NoError(t, err)
		time.Sleep(parentJob.InternalData.RetryInterval * 2)
		err = usecase.StopChainSessions(ctx, parentJob.ChainUUID)
		assert.NoError(t, err)
	})
	
	t.Run("should start and exit with errors after running retry session fails", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		parentJob.InternalData.RetryInterval = time.Second

		retrySessionState.EXPECT().Add(gomock.Any(), parentJob.UUID, parentJob)
		apiClient.EXPECT().SearchJob(gomock.Any(), gomock.Any()).Return([]*api.JobResponse{}, nil)
		sendRetryJobUseCase.EXPECT().Execute(gomock.Any(), parentJob, parentJob.UUID, 0).Return("", expectedErr)
		retrySessionState.EXPECT().Remove(gomock.Any(), parentJob.UUID)
		
		err := usecase.StartSession(ctx, parentJob)
		require.NoError(t, err)
		time.Sleep(parentJob.InternalData.RetryInterval*2)
		err = usecase.StopSession(ctx, parentJob.UUID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
	
	t.Run("should fail to stop not existing session", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		err := usecase.StopSession(ctx, parentJob.UUID)
		assert.Error(t, err)
		assert.True(t, errors.IsNotFoundError(err))
	})
	
	t.Run("should fail if add session fails", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		parentJob.InternalData.RetryInterval = time.Second

		retrySessionState.EXPECT().Add(gomock.Any(), parentJob.UUID, parentJob).Return(expectedErr)
		
		err := usecase.StartSession(ctx, parentJob)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
	
	t.Run("should skip new session if same job already have an open one", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		parentJob.InternalData.RetryInterval = time.Second*10

		retrySessionState.EXPECT().Add(gomock.Any(), parentJob.UUID, parentJob)
		apiClient.EXPECT().SearchJob(gomock.Any(), gomock.Any()).Return([]*api.JobResponse{}, nil)
		retrySessionState.EXPECT().Remove(gomock.Any(), parentJob.UUID)
		
		err := usecase.StartSession(ctx, parentJob)
		require.NoError(t, err)
		err = usecase.StartSession(ctx, parentJob)
		require.Error(t, err)
		assert.True(t, errors.IsAlreadyExistsError(err))
		time.Sleep(time.Second)
		err = usecase.StopSession(ctx, parentJob.UUID)
		require.NoError(t, err)
	})

}
