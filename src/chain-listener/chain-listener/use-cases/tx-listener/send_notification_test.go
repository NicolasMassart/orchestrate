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
	"github.com/consensys/orchestrate/src/entities/testdata"
	mock2 "github.com/consensys/orchestrate/src/infra/broker/sarama/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendNotification_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apiClient := mock.NewMockOrchestrateClient(ctrl)
	logger := log.NewLogger()
	producer := mock2.NewMockSyncProducer()

	topic := "tx-decoded-topic"
	expectedErr := fmt.Errorf("expected_err")

	contractName := "contractName"
	contractTag := "latest"
	usecase := SendNotificationUseCase(apiClient, producer, topic, logger)

	t.Run("should attach contract name and tag successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Receipt = testdata2.FakeReceipt()
		contractAddress := testdata.FakeAddress()
		job.Receipt.ContractAddress = contractAddress.String()

		apiClient.EXPECT().SearchContract(gomock.Any(), &types.SearchContractRequest{
			Address: contractAddress,
		}).Return(&types.ContractResponse{
			Name: contractName,
			Tag:  contractTag,
		}, nil)

		err := usecase.Execute(ctx, job)
		assert.NoError(t, err)
		assert.Equal(t, contractName, job.Receipt.ContractName)
		assert.Equal(t, contractTag, job.Receipt.ContractTag)
	})

	t.Run("should attach contract name and tag from logs successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Receipt = testdata2.FakeReceipt()
		contractAddress := testdata.FakeAddress()
		job.Receipt.Logs[0].Address = contractAddress.String()

		apiClient.EXPECT().SearchContract(gomock.Any(), &types.SearchContractRequest{
			Address: contractAddress,
		}).Return(&types.ContractResponse{
			Name: contractName,
			Tag:  contractTag,
		}, nil)

		err := usecase.Execute(ctx, job)
		assert.NoError(t, err)
		assert.Equal(t, contractName, job.Receipt.ContractName)
		assert.Equal(t, contractTag, job.Receipt.ContractTag)
	})

	t.Run("should not fail if there is not matched contract", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Receipt = testdata2.FakeReceipt()
		contractAddress := testdata.FakeAddress()
		job.Receipt.ContractAddress = contractAddress.String()

		apiClient.EXPECT().SearchContract(gomock.Any(), &types.SearchContractRequest{
			Address: contractAddress,
		}).Return(nil, errors.NotFoundError(""))

		err := usecase.Execute(ctx, job)
		assert.NoError(t, err)
		assert.Empty(t, job.Receipt.ContractName)
		assert.Empty(t, job.Receipt.ContractTag)
	})

	t.Run("should not fail if there is not contract address", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Receipt = testdata2.FakeReceipt()

		err := usecase.Execute(ctx, job)
		assert.NoError(t, err)
	})

	t.Run("should fail if client fails to retrieve contract data", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Receipt = testdata2.FakeReceipt()
		contractAddress := testdata.FakeAddress()
		job.Receipt.ContractAddress = contractAddress.String()

		apiClient.EXPECT().SearchContract(gomock.Any(), &types.SearchContractRequest{
			Address: contractAddress,
		}).Return(nil, expectedErr)

		err := usecase.Execute(ctx, job)
		require.Error(t, err)
		assert.True(t, errors.IsDependencyFailureError(err))
	})
}
