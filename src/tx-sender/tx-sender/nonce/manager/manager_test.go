// +build unit

package manager

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mock2 "github.com/consensys/orchestrate/src/infra/ethclient/mock"
	"github.com/consensys/orchestrate/src/tx-sender/store/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNonceManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ec := mock2.NewMockMultiClient(ctrl)
	ns := mock.NewMockNonceSender(ctrl)
	chainRegistryURL := "http://chain-registry:8081"
	rt := mock.NewMockRecoveryTracker(ctrl)
	maxRecovery := uint64(2)

	manager := NewNonceManager(ec, ns, rt, chainRegistryURL, maxRecovery)

	t.Run("should fetch nonce from chain successfully if no nonce is set", func(t *testing.T) {
		ctx := context.Background()
		job := testdata.FakeJob()
		expectedNonce := uint64(1)

		ns.EXPECT().GetLastSent(job.PartitionKey()).Return(uint64(0), errors.NotFoundError("error"))

		url := client.GetProxyURL(chainRegistryURL, job.ChainUUID)
		ec.EXPECT().PendingNonceAt(ctx, url, *job.Transaction.From).Return(expectedNonce, nil)

		nonce, err := manager.GetNonce(ctx, job)
		assert.NoError(t, err)
		assert.Equal(t, expectedNonce, nonce)
	})

	t.Run("should retrieve nonce from NonceSender successfully", func(t *testing.T) {
		ctx := context.Background()
		job := testdata.FakeJob()
		expectedNonce := uint64(2)

		ns.EXPECT().GetLastSent(job.PartitionKey()).Return(expectedNonce-1, nil)

		nonce, err := manager.GetNonce(ctx, job)
		assert.NoError(t, err)
		assert.Equal(t, expectedNonce, nonce)
	})

	t.Run("should return error if NonceSender fails", func(t *testing.T) {
		ctx := context.Background()
		job := testdata.FakeJob()

		expectedErr := errors.InvalidNonceWarning("invalid error")
		ns.EXPECT().GetLastSent(job.PartitionKey()).Return(uint64(0), expectedErr)

		_, err := manager.GetNonce(ctx, job)
		assert.Equal(t, err, expectedErr)
	})

	t.Run("should increment nonce successfully", func(t *testing.T) {
		ctx := context.Background()
		job := testdata.FakeJob()
		expectedNonce := uint64(1)
		job.Transaction.Nonce = utils.ToPtr(expectedNonce).(*uint64)

		ns.EXPECT().GetLastSent(job.PartitionKey()).Return(uint64(0), nil)
		ns.EXPECT().SetLastSent(job.PartitionKey(), expectedNonce).Return(nil)
		rt.EXPECT().Recovered(job.UUID)

		err := manager.IncrementNonce(ctx, job)
		assert.NoError(t, err)
	})

	t.Run("should increment nonce consecutively successfully", func(t *testing.T) {
		ctx := context.Background()
		job := testdata.FakeJob()
		expectedNonce := uint64(1)
		job.Transaction.Nonce = utils.ToPtr(expectedNonce).(*uint64)

		ns.EXPECT().GetLastSent(job.PartitionKey()).Return(uint64(0), nil)
		ns.EXPECT().SetLastSent(job.PartitionKey(), expectedNonce).Return(nil)
		rt.EXPECT().Recovered(job.UUID)

		err := manager.IncrementNonce(ctx, job)
		assert.NoError(t, err)
	})

	t.Run("should not increment nonce when no consecutively", func(t *testing.T) {
		ctx := context.Background()
		job := testdata.FakeJob()
		expectedNonce := uint64(2)
		job.Transaction.Nonce = utils.ToPtr(expectedNonce).(*uint64)

		ns.EXPECT().GetLastSent(job.PartitionKey()).Return(uint64(0), nil)
		rt.EXPECT().Recovered(job.UUID)

		err := manager.IncrementNonce(ctx, job)
		assert.NoError(t, err)
	})

	t.Run("should clean nonce if nonce too low error", func(t *testing.T) {
		ctx := context.Background()
		job := testdata.FakeJob()
		expectedNonce := uint64(1)
		job.Transaction.Nonce = utils.ToPtr(expectedNonce + 1).(*uint64)

		jobErr := errors.InvalidNonceWarning("nonce too low")
		ns.EXPECT().GetLastSent(job.PartitionKey()).Return(uint64(1), nil)
		ns.EXPECT().DeleteLastSent(job.PartitionKey()).Return(nil)
		rt.EXPECT().Recovering(job.UUID).Return(uint64(0))
		rt.EXPECT().Recover(job.UUID)

		err := manager.CleanNonce(ctx, job, jobErr)
		assert.True(t, errors.IsInvalidNonceWarning(err))
		assert.Empty(t, job.Transaction.Nonce)
	})

	t.Run("should not clean nonce if nonce too low error when it does not match nonce storage", func(t *testing.T) {
		ctx := context.Background()
		job := testdata.FakeJob()
		expectedNonce := uint64(1)
		job.Transaction.Nonce = utils.ToPtr(expectedNonce).(*uint64)

		jobErr := errors.InvalidNonceWarning("nonce too low")
		ns.EXPECT().GetLastSent(job.PartitionKey()).Return(expectedNonce, nil)
		rt.EXPECT().Recovering(job.UUID).Return(uint64(0))
		rt.EXPECT().Recover(job.UUID)

		err := manager.CleanNonce(ctx, job, jobErr)
		assert.True(t, errors.IsInvalidNonceWarning(err))
		assert.Empty(t, job.Transaction.Nonce)
	})

	t.Run("should do nothing if if nonce too low error", func(t *testing.T) {
		ctx := context.Background()
		job := testdata.FakeJob()
		expectedNonce := uint64(1)
		job.Transaction.Nonce = utils.ToPtr(expectedNonce).(*uint64)

		jobErr := errors.InvalidNonceWarning("internal error")

		err := manager.CleanNonce(ctx, job, jobErr)
		assert.NoError(t, err)
		assert.NotEmpty(t, job.Transaction.Nonce)
	})
}
