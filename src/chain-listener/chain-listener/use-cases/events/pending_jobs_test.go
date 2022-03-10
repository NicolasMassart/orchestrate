// +build unit

package events

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/sdk/client/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	proto "github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases/mocks"
	storemocks "github.com/consensys/orchestrate/src/chain-listener/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mock2 "github.com/consensys/orchestrate/src/infra/ethclient/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPendingJobs_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apiClient := mock.NewMockOrchestrateClient(ctrl)
	ethClient := mock2.NewMockMultiClient(ctrl)
	pendingJobStore := storemocks.NewMockPendingJob(ctrl)
	retryJobSessionManager := mocks.NewMockRetryJobSessionManager(ctrl)
	notifyMinedJobUC := mocks.NewMockNotifyMinedJob(ctrl)
	logger := log.NewLogger()

	proxyURL := "http://api"
	expectedErr := fmt.Errorf("expected_err")
	
	usecase := PendingJobUseCase(apiClient, ethClient, notifyMinedJobUC, retryJobSessionManager, pendingJobStore, logger)

	t.Run("should store pending job and start retry session", func(t *testing.T) {
		job := testdata.FakeJob()
		job.InternalData.RetryInterval = time.Second
		
		pendingJobStore.EXPECT().GetJobUUID(gomock.Any(), job.UUID).Return(nil, nil)
		apiClient.EXPECT().ChainProxyURL(job.ChainUUID).Return(proxyURL)
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(nil, nil)
		
		pendingJobStore.EXPECT().Add(gomock.Any(), job).Return(nil)
		retryJobSessionManager.EXPECT().StartSession(gomock.Any(), job).Return(nil)

		err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
	})
	
	t.Run("should store pending job without starting new retry session", func(t *testing.T) {
		job := testdata.FakeJob()
		
		pendingJobStore.EXPECT().GetJobUUID(gomock.Any(), job.UUID).Return(nil, nil)
		apiClient.EXPECT().ChainProxyURL(job.ChainUUID).Return(proxyURL)
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(nil, nil)
		
		pendingJobStore.EXPECT().Add(gomock.Any(), job).Return(nil)

		err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
	})
	
	t.Run("should store trigger notify mined job use case if receipt is available", func(t *testing.T) {
		job := testdata.FakeJob()
		
		pendingJobStore.EXPECT().GetJobUUID(gomock.Any(), job.UUID).Return(nil, nil)
		apiClient.EXPECT().ChainProxyURL(job.ChainUUID).Return(proxyURL)
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(&proto.Receipt{}, nil)
		notifyMinedJobUC.EXPECT().Execute(gomock.Any(), job).Return(nil)
		
		err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
	})
	
	t.Run("should not do anything if job already exist with same tx hash", func(t *testing.T) {
		job := testdata.FakeJob()
		
		pendingJobStore.EXPECT().GetJobUUID(gomock.Any(), job.UUID).Return(job, nil)
		
		err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
	})
	
	t.Run("should rerun flow if job already exist but with different tx hash", func(t *testing.T) {
		job := testdata.FakeJob()
		job2 := testdata.FakeJob()
		
		pendingJobStore.EXPECT().GetJobUUID(gomock.Any(), job.UUID).Return(job2, nil)
		apiClient.EXPECT().ChainProxyURL(job.ChainUUID).Return(proxyURL)
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(nil, nil)
		pendingJobStore.EXPECT().Update(gomock.Any(), job).Return(nil)
		
		err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
	})
	
	t.Run("should fail if notify mined job use case fails", func(t *testing.T) {
		job := testdata.FakeJob()
		
		pendingJobStore.EXPECT().GetJobUUID(gomock.Any(), job.UUID).Return(nil, nil)
		apiClient.EXPECT().ChainProxyURL(job.ChainUUID).Return(proxyURL)
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(&proto.Receipt{}, nil)
		notifyMinedJobUC.EXPECT().Execute(gomock.Any(), job).Return(expectedErr)
		
		err := usecase.Execute(ctx, job)

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
	
	
	t.Run("should fail if start session fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.InternalData.RetryInterval = time.Second
		
		pendingJobStore.EXPECT().GetJobUUID(gomock.Any(), job.UUID).Return(nil, nil)
		apiClient.EXPECT().ChainProxyURL(job.ChainUUID).Return(proxyURL)
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(nil, nil)
		
		pendingJobStore.EXPECT().Add(gomock.Any(), job).Return(nil)
		retryJobSessionManager.EXPECT().StartSession(gomock.Any(), job).Return(expectedErr)

		err := usecase.Execute(ctx, job)

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
