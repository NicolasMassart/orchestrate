// +build unit

package txlistener

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/client/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	testdata2 "github.com/consensys/orchestrate/pkg/types/ethereum/testdata"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases/mocks"
	storemocks "github.com/consensys/orchestrate/src/chain-listener/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mock2 "github.com/consensys/orchestrate/src/infra/ethclient/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifyMinedJob_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apiClient := mock.NewMockOrchestrateClient(ctrl)
	ethClient := mock2.NewMockMultiClient(ctrl)
	sendNotification := mocks.NewMockSendNotification(ctrl)
	registerDeployedContract := mocks.NewMockRegisterDeployedContract(ctrl)
	chainState := storemocks.NewMockChain(ctrl)
	logger := log.NewLogger()

	proxyURL := "http://api"
	apiClient.EXPECT().ChainProxyURL(gomock.Any()).AnyTimes().Return(proxyURL)
	expectedErr := fmt.Errorf("expected_err")
	
	chain := testdata.FakeChain()
	usecase := NotifyMinedJobUseCase(apiClient, ethClient, sendNotification, registerDeployedContract, chainState, logger)

	t.Run("should handle mined job successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = chain.UUID
		receipt := testdata2.FakeReceipt()
		receipt.ContractAddress = testdata.FakeAddress().String()
		
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(receipt, nil)
		chainState.EXPECT().Get(gomock.Any(), job.ChainUUID).Return(chain, nil)
		apiClient.EXPECT().GetContractEvents(gomock.Any(), receipt.Logs[0].Address, chain.ChainID.String(), gomock.Any()).
			Return(&types.GetContractEventsBySignHashResponse{}, nil)
		
		registerDeployedContract.EXPECT().Execute(gomock.Any(), job).Return(nil)
		apiClient.EXPECT().UpdateJob(gomock.Any(), job.UUID, gomock.Any()).Return(&types.JobResponse{}, nil)
		sendNotification.EXPECT().Execute(gomock.Any(), job).Return(nil)
		
		err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
	})
	
	t.Run("should fail to handle mined job if notifies fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = chain.UUID
		receipt := testdata2.FakeReceipt()
		receipt.Logs = nil
		
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(receipt, nil)
		chainState.EXPECT().Get(gomock.Any(), job.ChainUUID).Return(chain, nil)
		apiClient.EXPECT().UpdateJob(gomock.Any(), job.UUID, gomock.Any()).Return(&types.JobResponse{}, nil)
		sendNotification.EXPECT().Execute(gomock.Any(), job).Return(expectedErr)
		
		err := usecase.Execute(ctx, job)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
	
	t.Run("should fail to handle mined job if update status fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = chain.UUID
		receipt := testdata2.FakeReceipt()
		
		ethClient.EXPECT().TransactionReceipt(gomock.Any(), proxyURL, *job.Transaction.Hash).Return(receipt, nil)
		chainState.EXPECT().Get(gomock.Any(), job.ChainUUID).Return(chain, nil)
		apiClient.EXPECT().GetContractEvents(gomock.Any(), receipt.Logs[0].Address, chain.ChainID.String(), gomock.Any()).
			Return(&types.GetContractEventsBySignHashResponse{}, nil)
		
		apiClient.EXPECT().UpdateJob(gomock.Any(), job.UUID, gomock.Any()).Return(&types.JobResponse{}, expectedErr)
		
		err := usecase.Execute(ctx, job)

		require.Error(t, err)
		assert.True(t, errors.IsDependencyFailureError(err))
	})
}