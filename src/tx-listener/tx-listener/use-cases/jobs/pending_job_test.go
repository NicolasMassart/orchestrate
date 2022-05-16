// +build unit

package jobs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/sdk/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	proto "github.com/consensys/orchestrate/pkg/types/ethereum"
	testdata3 "github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mock2 "github.com/consensys/orchestrate/src/infra/ethclient/mock"
	storemocks "github.com/consensys/orchestrate/src/tx-listener/store/mocks"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/mocks"
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
	pendingJobState := storemocks.NewMockPendingJob(ctrl)
	messengerState := storemocks.NewMockMessage(ctrl)
	minedJobUC := mocks.NewMockMinedJob(ctrl)
	logger := log.NewLogger()

	proxyURL := "http://api"
	expectedErr := fmt.Errorf("expected_err")
	
	usecase := PendingJob(apiClient, ethClient, minedJobUC, pendingJobState, messengerState, logger)

	t.Run("should store pending job and start retry session", func(t *testing.T) {
		job := testdata.FakeJob()
		job.InternalData.RetryInterval = time.Second
		msg := testdata3.NewFakeJobUpdateMessage(job.UUID)
		
		pendingJobState.EXPECT().GetJobUUID(gomock.Any(), job.UUID).Return(nil, nil)
		apiClient.EXPECT().ChainProxyURL(job.ChainUUID).Return(proxyURL)
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(nil, nil)
		messengerState.EXPECT().AddJobMessage(gomock.Any(), job.UUID, msg).Return(nil)
		pendingJobState.EXPECT().Add(gomock.Any(), job).Return(nil)

		err := usecase.Execute(ctx, job, msg)

		assert.NoError(t, err)
	})
	
	t.Run("should store trigger notify mined job use case if receipt is available", func(t *testing.T) {
		job := testdata.FakeJob()
		msg := testdata3.NewFakeJobUpdateMessage(job.UUID)
		
		pendingJobState.EXPECT().GetJobUUID(gomock.Any(), job.UUID).Return(nil, nil)
		apiClient.EXPECT().ChainProxyURL(job.ChainUUID).Return(proxyURL)
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(&proto.Receipt{}, nil)
		minedJobUC.EXPECT().Execute(gomock.Any(), job).Return(nil)
		
		err := usecase.Execute(ctx, job, msg)
	
		assert.NoError(t, err)
	})
	
	t.Run("should not do anything if job already exist with same tx hash", func(t *testing.T) {
		job := testdata.FakeJob()
		msg := testdata3.NewFakeJobUpdateMessage(job.UUID)
		
		pendingJobState.EXPECT().GetJobUUID(gomock.Any(), job.UUID).Return(job, nil)
		
		err := usecase.Execute(ctx, job, msg)
	
		assert.NoError(t, err)
	})
	
	t.Run("should rerun flow if job already exist but with different tx hash", func(t *testing.T) {
		job := testdata.FakeJob()
		job2 := testdata.FakeJob()
		msg := testdata3.NewFakeJobUpdateMessage(job.UUID)
		
		pendingJobState.EXPECT().GetJobUUID(gomock.Any(), job.UUID).Return(job2, nil)
		apiClient.EXPECT().ChainProxyURL(job.ChainUUID).Return(proxyURL)
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(nil, nil)
		pendingJobState.EXPECT().Update(gomock.Any(), job).Return(nil)
		
		err := usecase.Execute(ctx, job, msg)
	
		assert.NoError(t, err)
	})
	
	t.Run("should fail if notify mined job use case fails", func(t *testing.T) {
		job := testdata.FakeJob()
		msg := testdata3.NewFakeJobUpdateMessage(job.UUID)
		
		pendingJobState.EXPECT().GetJobUUID(gomock.Any(), job.UUID).Return(nil, nil)
		apiClient.EXPECT().ChainProxyURL(job.ChainUUID).Return(proxyURL)
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(&proto.Receipt{}, nil)
		minedJobUC.EXPECT().Execute(gomock.Any(), job).Return(expectedErr)
		
		err := usecase.Execute(ctx, job, msg)
	
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
