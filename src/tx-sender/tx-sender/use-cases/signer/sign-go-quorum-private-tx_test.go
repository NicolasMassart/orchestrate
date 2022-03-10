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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignQuorumPrivateTransaction_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyManagerClient := qkmmock.NewMockKeyManagerClient(ctrl)
	ctx := context.Background()

	usecase := NewSignGoQuorumPrivateTransactionUseCase(mockKeyManagerClient)

	signedRaw := utils.StringToHexBytes("0xf851018227108252088082c35080820713a09a0a890215ea6e79d06f9665297996ab967db117f36c2090d6d6ead5a2d32d52a065bc4bc766b5a833cb58b3319e44e952487559b9b939cb5268c0409398214c8b")

	t.Run("should execute use case successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		mockKeyManagerClient.EXPECT().SignQuorumPrivateTransaction(gomock.Any(), job.InternalData.StoreID, job.Transaction.From.String(), 
			gomock.AssignableToTypeOf(&types.SignQuorumPrivateTransactionRequest{})).Return(signedRaw.String(), nil)

		raw, txHash, err := usecase.Execute(ctx, job)

		require.NoError(t, err)
		assert.Equal(t, signedRaw.String(), raw.String())
		assert.Equal(t, "0x2c8bdef96dca7d037618fcf799a4bbfec6a6e1299b27dbc5d7cd79594e31ee54", txHash.String())
	})

	t.Run("should execute use case successfully for deployment transactions", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Transaction.To = nil
		mockKeyManagerClient.EXPECT().SignQuorumPrivateTransaction(gomock.Any(), job.InternalData.StoreID, job.Transaction.From.String(), 
			gomock.AssignableToTypeOf(&types.SignQuorumPrivateTransactionRequest{})).Return(signedRaw.String(), nil)

		raw, txHash, err := usecase.Execute(ctx, job)

		require.NoError(t, err)
		assert.Equal(t, signedRaw.String(), raw.String())
		assert.Equal(t, "0x2c8bdef96dca7d037618fcf799a4bbfec6a6e1299b27dbc5d7cd79594e31ee54", txHash.String())
	})

	t.Run("should execute use case successfully for one time key transactions", func(t *testing.T) {
		job := testdata.FakeJob()
		job.InternalData.OneTimeKey = true

		raw, txHash, err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
		assert.NotEmpty(t, raw)
		assert.NotEmpty(t, txHash)
	})

	t.Run("should fail with same error if ETHSignQuorumPrivateTransaction fails", func(t *testing.T) {
		expectedErr := errors.InvalidFormatError("error")
		job := testdata.FakeJob()
		mockKeyManagerClient.EXPECT().SignQuorumPrivateTransaction(gomock.Any(), job.InternalData.StoreID, gomock.Any(), gomock.Any()).
			Return("", expectedErr)

		raw, txHash, err := usecase.Execute(ctx, job)

		assert.True(t, errors.IsDependencyFailureError(err))
		assert.Empty(t, raw)
		assert.Empty(t, txHash)
	})
}
