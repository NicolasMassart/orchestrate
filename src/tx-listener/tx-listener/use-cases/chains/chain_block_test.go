// +build unit

package chains
// 
// import (
// 	"context"
// 	"fmt"
// 	"testing"
// 
// 	"github.com/consensys/orchestrate/pkg/errors"
// 	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
// 	storemocks "github.com/consensys/orchestrate/src/tx-listener/store/mocks"
// 	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/mocks"
// 	ethcommon "github.com/ethereum/go-ethereum/common"
// 	"github.com/stretchr/testify/require"
// 
// 	"github.com/consensys/orchestrate/src/entities/testdata"
// 	"github.com/golang/mock/gomock"
// 	"github.com/stretchr/testify/assert"
// )
// 
// func TestChainBlock_Execute(t *testing.T) {
// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
// 
// 	pendingJobStore := storemocks.NewMockPendingJob(ctrl)
// 	retrySessionsStore := storemocks.NewMockRetrySessions(ctrl)
// 	retryJobSessionManager := mocks.NewMockRetryJobSessionManager(ctrl)
// 	notifyMinedJobUC := mocks.NewMockNotifyMinedJob(ctrl)
// 	logger := log.NewLogger()
// 
// 	chain := testdata.FakeChain()
// 	expectedErr := fmt.Errorf("expected_err")
// 
// 	blockNumber := uint64(1)
// 	usecase := NewChainBlockUseCase(notifyMinedJobUC, pendingJobStore, retrySessionsStore, logger)
// 
// 	t.Run("should handle mined jobs successfully", func(t *testing.T) {
// 		jobOne := testdata.FakeJob()
// 		jobTwo := testdata.FakeJob()
// 		txHashes := []*ethcommon.Hash{testdata.FakeTxHash(), testdata.FakeTxHash()}
// 
// 		pendingJobStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[0]).Return(jobOne, nil)
// 		pendingJobStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[1]).Return(jobTwo, nil)
// 
// 		notifyMinedJobUC.EXPECT().Execute(gomock.Any(), jobOne).Return(nil)
// 		notifyMinedJobUC.EXPECT().Execute(gomock.Any(), jobTwo).Return(nil)
// 
// 		pendingJobStore.EXPECT().Remove(gomock.Any(), jobOne.UUID).Return(nil)
// 		pendingJobStore.EXPECT().Remove(gomock.Any(), jobTwo.UUID).Return(nil)
// 
// 		sessIDOne := "sessIDOne"
// 		sessIDTwo := "sessIDTwo"
// 		retrySessionsStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[0]).Return(sessIDOne, nil)
// 		retrySessionsStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[1]).Return(sessIDTwo, nil)
// 
// 		retryJobSessionManager.EXPECT().StopSession(gomock.Any(), sessIDOne).Return(nil)
// 		retryJobSessionManager.EXPECT().StopSession(gomock.Any(), sessIDTwo).Return(nil)
// 
// 		err := usecase.Execute(ctx, chain.UUID, blockNumber, txHashes)
// 
// 		assert.NoError(t, err)
// 	})
// 
// 	t.Run("should not fail if there is not retry sessions", func(t *testing.T) {
// 		jobOne := testdata.FakeJob()
// 		txHashes := []*ethcommon.Hash{testdata.FakeTxHash()}
// 
// 		pendingJobStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[0]).Return(jobOne, nil)
// 
// 		notifyMinedJobUC.EXPECT().Execute(gomock.Any(), jobOne).Return(nil)
// 
// 		pendingJobStore.EXPECT().Remove(gomock.Any(), jobOne.UUID).Return(nil)
// 
// 		sessIDOne := "sessIDOne"
// 		retrySessionsStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[0]).
// 			Return(sessIDOne, errors.NotFoundError(""))
// 
// 		err := usecase.Execute(ctx, chain.UUID, blockNumber, txHashes)
// 
// 		assert.NoError(t, err)
// 	})
// 
// 	t.Run("should fail if stop retry session fails", func(t *testing.T) {
// 		txHashes := []*ethcommon.Hash{testdata.FakeTxHash()}
// 
// 		sessIDOne := "sessIDOne"
// 		jobOne := testdata.FakeJob()
// 		pendingJobStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[0]).Return(jobOne, nil)
// 		notifyMinedJobUC.EXPECT().Execute(gomock.Any(), jobOne).Return(nil)
// 		pendingJobStore.EXPECT().Remove(gomock.Any(), jobOne.UUID).Return(nil)
// 		retrySessionsStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[0]).Return(sessIDOne, nil)
// 		retryJobSessionManager.EXPECT().StopSession(gomock.Any(), sessIDOne).Return(expectedErr)
// 
// 		err := usecase.Execute(ctx, chain.UUID, blockNumber, txHashes)
// 
// 		require.Error(t, err)
// 		assert.Equal(t, expectedErr, err)
// 	})
// 
// 	t.Run("should fail if it fails to send notification", func(t *testing.T) {
// 		jobOne := testdata.FakeJob()
// 		txHashes := []*ethcommon.Hash{testdata.FakeTxHash()}
// 
// 		sessIDOne := "sessIDOne"
// 		retrySessionsStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[0]).Return(sessIDOne, nil)
// 		retryJobSessionManager.EXPECT().StopSession(gomock.Any(), sessIDOne).Return(nil)
// 
// 		pendingJobStore.EXPECT().GetByTxHash(gomock.Any(), chain.UUID, txHashes[0]).Return(jobOne, nil)
// 		notifyMinedJobUC.EXPECT().Execute(gomock.Any(), jobOne).Return(expectedErr)
// 
// 		err := usecase.Execute(ctx, chain.UUID, blockNumber, txHashes)
// 
// 		require.Error(t, err)
// 		assert.Equal(t, expectedErr, err)
// 	})
// }
