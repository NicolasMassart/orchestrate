// +build unit

package jobs

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/sdk/client/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/types"
	apitestdata "github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateChildJob_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	initialGasPrice := utils.StringBigIntToHex("1000000000")
	ctx := context.Background()

	logger := log.NewLogger()
	apiClient := mock.NewMockOrchestrateClient(ctrl)

	usecase := RetrySessionJobUseCase(apiClient, logger)

	t.Run("should create a new child job successfully", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		childJobResponse := apitestdata.FakeJobResponse()
		parentJob.Transaction.GasPrice = initialGasPrice
		parentJob.InternalData.GasPriceIncrement = 0.1
		parentJob.InternalData.GasPriceLimit = 0.2

		apiClient.EXPECT().CreateJob(gomock.Any(), gomock.Any()).Return(childJobResponse, nil)
		apiClient.EXPECT().StartJob(gomock.Any(), childJobResponse.UUID).Return(nil)

		childJobUUID, err := usecase.Execute(ctx, parentJob, parentJob.UUID, 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, childJobUUID)
	})

	t.Run("should resend job transaction if the parent job status is PENDING with not gas increment", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		parentJob.Transaction.GasPrice = initialGasPrice

		apiClient.EXPECT().ResendJobTx(gomock.Any(), parentJob.UUID).Return(nil)

		childJobUUID, err := usecase.Execute(ctx, parentJob, parentJob.UUID, 0)
		assert.NoError(t, err)
		assert.Equal(t, childJobUUID, parentJob.UUID)
	})

	t.Run("should resend job transaction last job if gas limit was reached", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		childJob := testdata.FakeJob()
		parentJob.Transaction.GasPrice = initialGasPrice
		parentJob.InternalData.GasPriceIncrement = 0.1
		parentJob.InternalData.GasPriceLimit = 0.2

		apiClient.EXPECT().ResendJobTx(gomock.Any(), childJob.UUID).Return(nil)

		childJobUUID, err := usecase.Execute(ctx, parentJob, childJob.UUID, 3)
		assert.NoError(t, err)
		assert.Equal(t, childJobUUID, parentJob.UUID)
	})

	t.Run("should send the same job if job is a raw transaction", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		parentJob.Transaction.Raw = utils.StringToHexBytes("0xAB")
		parentJob.Type = entities.EthereumRawTransaction

		apiClient.EXPECT().ResendJobTx(gomock.Any(), parentJob.UUID).Return(nil)

		childJobUUID, err := usecase.Execute(ctx, parentJob, parentJob.UUID, 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, childJobUUID)
	})

	t.Run("should create a new child job by increasing the gasPrice by Increment (legacyTx)", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		childJob := testdata.FakeJob()
		childJobResponse := apitestdata.FakeJobResponse()

		parentJobResponse := apitestdata.FakeJobResponse()
		parentJob.Transaction.Nonce = utils.ToPtr(uint64(1)).(*uint64)
		parentJob.Transaction.TransactionType = entities.LegacyTxType
		parentJob.Transaction.GasPrice = initialGasPrice
		parentJob.InternalData.GasPriceIncrement = 0.06
		parentJob.InternalData.GasPriceLimit = 0.12

		apiClient.EXPECT().CreateJob(gomock.Any(), gomock.Any()).
			DoAndReturn(func(timeoutCtx context.Context, req *types.CreateJobRequest) (*types.JobResponse, error) {
				assert.Equal(t, "1120000000", req.Transaction.GasPrice.ToInt().String())
				assert.Equal(t, parentJobResponse.Transaction.Nonce, req.Transaction.Nonce)
				return childJobResponse, nil
			})
		apiClient.EXPECT().StartJob(gomock.Any(), childJobResponse.UUID).Return(nil)

		childJobUUID, err := usecase.Execute(ctx, parentJob, childJob.UUID, 1)
		assert.NoError(t, err)
		assert.NotEmpty(t, childJobUUID)
	})
	
	t.Run("should create a new child job by increasing the gasPrice by Increment (dynamicTx)", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		childJob := testdata.FakeJob()
		childJobResponse := apitestdata.FakeJobResponse()

		parentJob.Transaction.Nonce = utils.ToPtr(uint64(1)).(*uint64)
		parentJob.Transaction.TransactionType = entities.DynamicFeeTxType
		parentJob.Transaction.GasTipCap = initialGasPrice
		parentJob.InternalData.GasPriceIncrement =0.06
		parentJob.InternalData.GasPriceLimit =0.12

		apiClient.EXPECT().CreateJob(gomock.Any(), gomock.Any()).
			DoAndReturn(func(timeoutCtx context.Context, req *types.CreateJobRequest) (*types.JobResponse, error) {
				assert.Equal(t, "1120000000", req.Transaction.GasTipCap.ToInt().String())
				assert.Equal(t, parentJob.Transaction.Nonce, req.Transaction.Nonce)
				return childJobResponse, nil
			})
		apiClient.EXPECT().StartJob(gomock.Any(), childJobResponse.UUID).Return(nil)

		childJobUUID, err := usecase.Execute(ctx, parentJob, childJob.UUID, 1)
		assert.NoError(t, err)
		assert.NotEmpty(t, childJobUUID)
	})

	t.Run("should create a new child job by increasing the gasPrice and not exceed the limit (legacyTx)", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		childJob := testdata.FakeJob()
		childJobResponse := apitestdata.FakeJobResponse()

		parentJob.Transaction.Nonce = utils.ToPtr(uint64(1)).(*uint64)
		parentJob.Transaction.TransactionType = entities.LegacyTxType
		parentJob.Transaction.GasPrice = initialGasPrice
		parentJob.InternalData.GasPriceIncrement = 0.06
		parentJob.InternalData.GasPriceLimit = 0.05

		apiClient.EXPECT().CreateJob(gomock.Any(), gomock.Any()).
			DoAndReturn(func(timeoutCtx context.Context, req *types.CreateJobRequest) (*types.JobResponse, error) {
				assert.Equal(t, "1050000000", req.Transaction.GasPrice.ToInt().String())
				assert.Equal(t, parentJob.Transaction.Nonce, req.Transaction.Nonce)
				return childJobResponse, nil
			})
		apiClient.EXPECT().StartJob(gomock.Any(), childJobResponse.UUID).Return(nil)

		childJobUUID, err := usecase.Execute(ctx, parentJob, childJob.UUID, 1)
		assert.NoError(t, err)
		assert.NotEmpty(t, childJobUUID)
	})
	
	t.Run("should create a new child job by increasing the gasPrice and not exceed the limit (dynamicTx)", func(t *testing.T) {
		parentJob := testdata.FakeJob()
		childJob := testdata.FakeJob()
		childJobResponse := apitestdata.FakeJobResponse()

		parentJob.Transaction.Nonce = utils.ToPtr(uint64(1)).(*uint64)
		parentJob.Transaction.TransactionType = entities.DynamicFeeTxType
		parentJob.Transaction.GasTipCap = initialGasPrice
		parentJob.InternalData.GasPriceIncrement =0.06
		parentJob.InternalData.GasPriceLimit =0.05
		
		apiClient.EXPECT().CreateJob(gomock.Any(), gomock.Any()).
			DoAndReturn(func(timeoutCtx context.Context, req *types.CreateJobRequest) (*types.JobResponse, error) {
				assert.Equal(t, "1050000000", req.Transaction.GasTipCap.ToInt().String())
				assert.Equal(t, parentJob.Transaction.Nonce, req.Transaction.Nonce)
				return childJobResponse, nil
			})
		apiClient.EXPECT().StartJob(gomock.Any(), childJobResponse.UUID).Return(nil)

		childJobUUID, err := usecase.Execute(ctx, parentJob, childJob.UUID, 1)
		assert.NoError(t, err)
		assert.NotEmpty(t, childJobUUID)
	})
}
