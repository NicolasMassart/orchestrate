// +build unit

package sender

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/client/mock"
	testdata2 "github.com/consensys/orchestrate/src/api/service/types/testdata"
	mock2 "github.com/consensys/orchestrate/src/infra/ethclient/mock"
	apitypes "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/orchestrate/pkg/utils"
	mocks2 "github.com/consensys/orchestrate/src/tx-sender/tx-sender/nonce/mocks"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases/mocks"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
)

func TestSendEEAPrivate_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signTx := mocks.NewMockSignEEATransactionUseCase(ctrl)
	ec := mock2.NewMockEEATransactionSender(ctrl)
	crafter := mocks.NewMockCraftTransactionUseCase(ctrl)
	jobClient := mock.NewMockJobClient(ctrl)
	nonceManager := mocks2.NewMockManager(ctrl)
	chainRegistryURL := "chainRegistryURL:8081"
	ctx := context.Background()
	raw := hexutil.MustDecode("0x1234")
	txHash := ethcommon.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000abc")

	usecase := NewSendEEAPrivateTxUseCase(signTx, crafter, ec, jobClient, chainRegistryURL, nonceManager)

	t.Run("should execute use case successfully", func(t *testing.T) {
		job := testdata.FakeJob()

		crafter.EXPECT().Execute(gomock.Any(), job).Return(nil)
		signTx.EXPECT().Execute(gomock.Any(), job).Return(raw, &txHash, nil)
		job.Transaction.Raw = raw
		job.Transaction.Hash = &txHash

		proxyURL := utils.GetProxyURL(chainRegistryURL, job.ChainUUID)
		ec.EXPECT().PrivDistributeRawTransaction(gomock.Any(), proxyURL, job.Transaction.Raw).Return(txHash, nil)
		nonceManager.EXPECT().IncrementNonce(gomock.Any(), job).Return(nil)
		jobClient.EXPECT().UpdateJob(gomock.Any(), job.UUID, gomock.Any()).DoAndReturn(func(ctx context.Context, jobUUID string, request *apitypes.UpdateJobRequest) (*apitypes.JobResponse, error) {
			assert.Equal(t, request.Status, entities.StatusStored)
			assert.Equal(t, request.Transaction.Hash, job.Transaction.Hash)
			assert.Equal(t, request.Transaction.Raw, job.Transaction.Raw)
			return testdata2.FakeJobResponse(), nil
		})

		err := usecase.Execute(ctx, job)
		assert.NoError(t, err)
		assert.Equal(t, job.Transaction.Hash.String(), txHash.String())
	})

	t.Run("should fail to execute use case if nonce checker fails", func(t *testing.T) {
		job := testdata.FakeJob()

		expectedErr := errors.NonceTooLowWarning("invalid nonce")
		crafter.EXPECT().Execute(gomock.Any(), job).Return(expectedErr)

		err := usecase.Execute(ctx, job)
		assert.Equal(t, err, expectedErr)
	})

	t.Run("should fail to execute use case if signer fails", func(t *testing.T) {
		job := testdata.FakeJob()

		crafter.EXPECT().Execute(gomock.Any(), job).Return(nil)

		expectedErr := errors.InternalError("internal error")
		signTx.EXPECT().Execute(gomock.Any(), job).Return(raw, &txHash, expectedErr)

		err := usecase.Execute(ctx, job)
		assert.Equal(t, err, expectedErr)
	})

	t.Run("should fail to execute use case if send transaction fails", func(t *testing.T) {
		job := testdata.FakeJob()

		crafter.EXPECT().Execute(gomock.Any(), job).Return(nil)
		signTx.EXPECT().Execute(gomock.Any(), job).Return(raw, &txHash, nil)
		job.Transaction.Raw = raw
		job.Transaction.Hash = &txHash

		expectedErr := errors.InternalError("internal error")
		proxyURL := utils.GetProxyURL(chainRegistryURL, job.ChainUUID)
		ec.EXPECT().PrivDistributeRawTransaction(gomock.Any(), proxyURL, job.Transaction.Raw).Return(ethcommon.HexToHash(""), expectedErr)
		nonceManager.EXPECT().CleanNonce(gomock.Any(), job, expectedErr).Return(nil)

		err := usecase.Execute(ctx, job)
		assert.Equal(t, err, expectedErr)
	})
}
