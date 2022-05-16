// +build unit

package chains

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities/testdata"
	storemocks "github.com/consensys/orchestrate/src/tx-listener/store/mocks"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/mocks"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainBlockTxs_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pendingJobStore := storemocks.NewMockPendingJob(ctrl)
	minedJobUC := mocks.NewMockMinedJob(ctrl)
	logger := log.NewLogger()

	chain := testdata.FakeChain()
	expectedErr := fmt.Errorf("expected_err")

	blockNumber := uint64(1)
	usecase := NewChainBlockTxsUseCase(minedJobUC, pendingJobStore, logger)

	t.Run("should handle mined jobs successfully", func(t *testing.T) {
		jobOne := testdata.FakeJob()
		jobTwo := testdata.FakeJob()
		txHashes := []*ethcommon.Hash{testdata.FakeTxHash(), testdata.FakeTxHash()}

		pendingJobStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[0]).Return(jobOne, nil)
		pendingJobStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[1]).Return(jobTwo, nil)

		minedJobUC.EXPECT().Execute(gomock.Any(), jobOne).Return(nil)
		minedJobUC.EXPECT().Execute(gomock.Any(), jobTwo).Return(nil)

		err := usecase.Execute(ctx, chain.UUID, blockNumber, txHashes)

		assert.NoError(t, err)
	})

	t.Run("should fail if it fails to update to mined", func(t *testing.T) {
		jobOne := testdata.FakeJob()
		txHashes := []*ethcommon.Hash{testdata.FakeTxHash()}

		pendingJobStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[0]).Return(jobOne, nil)
		minedJobUC.EXPECT().Execute(gomock.Any(), jobOne).Return(expectedErr)

		err := usecase.Execute(ctx, chain.UUID, blockNumber, txHashes)

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
