// +build unit

package jobs

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	testdata2 "github.com/consensys/orchestrate/pkg/types/ethereum/testdata"
	testdata3 "github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mock2 "github.com/consensys/orchestrate/src/infra/ethclient/mock"
	mocks3 "github.com/consensys/orchestrate/src/tx-listener/store/mocks"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifyMinedJob_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	completedJobUC := mocks.NewMockCompletedJob(ctrl)
	messengerAPI := mock.NewMockMessengerAPI(ctrl)
	chainProxyClient := mock.NewMockChainProxyClient(ctrl)
	ethClient := mock2.NewMockMultiClient(ctrl)
	registerDeployedContract := mocks.NewMockRegisterDeployedContract(ctrl)
	pendingJobState := mocks3.NewMockPendingJob(ctrl)
	logger := log.NewLogger()

	proxyURL := "http://api"
	chainProxyClient.EXPECT().ChainProxyURL(gomock.Any()).AnyTimes().Return(proxyURL)
	expectedErr := fmt.Errorf("expected_err")

	chain := testdata.FakeChain()
	usecase := MinedJobUseCase(messengerAPI, chainProxyClient, ethClient, completedJobUC, registerDeployedContract,
		pendingJobState, logger)

	t.Run("should handle mined job successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = chain.UUID
		receipt := testdata2.FakeReceipt()
		receipt.ContractAddress = testdata.FakeAddress().String()

		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(receipt, nil)

		registerDeployedContract.EXPECT().Execute(gomock.Any(), job).Return(nil)

		messengerAPI.EXPECT().JobUpdateMessage(gomock.Any(),
			testdata3.MinedJobMessageRequestMatcher(job.UUID, receipt), gomock.Any()).
			Return(nil)

		completedJobUC.EXPECT().Execute(gomock.Any(), job).Return(nil)
		err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
	})

	t.Run("should fail to handle mined job if update status fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = chain.UUID
		receipt := testdata2.FakeReceipt()
		
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(receipt, nil)
		
		messengerAPI.EXPECT().JobUpdateMessage(gomock.Any(),
			testdata3.MinedJobMessageRequestMatcher(job.UUID, receipt), gomock.Any()).
			Return(expectedErr)
		
		err := usecase.Execute(ctx, job)
	
		require.Error(t, err)
		assert.True(t, errors.IsDependencyFailureError(err))
	})
}
