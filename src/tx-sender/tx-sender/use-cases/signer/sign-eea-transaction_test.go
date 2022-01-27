// +build unit

package signer

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/orchestrate/pkg/utils"
	qkmmock "github.com/consensys/quorum-key-manager/pkg/client/mock"
	"github.com/consensys/quorum-key-manager/src/stores/api/types"
	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSignEEATransaction_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyManagerClient := qkmmock.NewMockKeyManagerClient(ctrl)
	ctx := context.Background()

	usecase := NewSignEEATransactionUseCase(mockKeyManagerClient)

	signedRaw := utils.StringToHexBytes("0xf8d501822710825208944fed1fc4144c223ae3c1553be203cdfcbd38c58182c35080820713a09a0a890215ea6e79d06f9665297996ab967db117f36c2090d6d6ead5a2d32d52a065bc4bc766b5a833cb58b3319e44e952487559b9b939cb5268c0409398214c8ba0035695b4cc4b0941e60551d7a19cf30603db5bfc23e5ac43a56f57f25f75486af842a0035695b4cc4b0941e60551d7a19cf30603db5bfc23e5ac43a56f57f25f75486aa0075695b4cc4b0941e60551d7a19cf30603db5bfc23e5ac43a56f57f25f75486a8a72657374726963746564")
	
	t.Run("should execute use case successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		mockKeyManagerClient.EXPECT().SignEEATransaction(gomock.Any(), job.InternalData.StoreID, job.Transaction.From.String(), 
			gomock.AssignableToTypeOf(&types.SignEEATransactionRequest{})).Return(signedRaw.String(), nil)

		raw, txHash, err := usecase.Execute(ctx, job)

		require.NoError(t, err)
		assert.Equal(t, signedRaw.String(), raw.String())
		assert.Empty(t, txHash)
	})

	t.Run("should execute use case successfully for deployment transactions", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Transaction.To = nil
		mockKeyManagerClient.EXPECT().SignEEATransaction(gomock.Any(), job.InternalData.StoreID, job.Transaction.From.String(), 
			gomock.AssignableToTypeOf(&types.SignEEATransactionRequest{})).Return(signedRaw.String(), nil)

		raw, txHash, err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
		assert.Equal(t, signedRaw.String(), raw.String())
		assert.Empty(t, txHash)
	})

	t.Run("should execute use case successfully for one time key transactions", func(t *testing.T) {
		job := testdata.FakeJob()
		job.InternalData.OneTimeKey = true
		job.Transaction.PrivateFrom = "A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="
		job.Transaction.PrivateFor = []string{"A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="}
		job.Transaction.TransactionType = types.LegacyTxType

		raw, txHash, err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
		assert.NotEmpty(t, raw)
		assert.Empty(t, txHash)
	})

	t.Run("should fail with same error if ETHSignEEATransaction fails", func(t *testing.T) {
		job := testdata.FakeJob()
		mockKeyManagerClient.EXPECT().SignEEATransaction(gomock.Any(), job.InternalData.StoreID, job.Transaction.From.String(), gomock.Any()).
			Return("", errors.InvalidFormatError("error"))

		raw, txHash, err := usecase.Execute(ctx, job)

		assert.True(t, errors.IsDependencyFailureError(err))
		assert.Empty(t, raw)
		assert.Empty(t, txHash)
	})
}
