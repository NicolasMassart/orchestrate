// +build unit
// +build !race

package chains

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/sdk/client/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	testdata2 "github.com/consensys/orchestrate/pkg/types/ethereum/testdata"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mock2 "github.com/consensys/orchestrate/src/infra/ethclient/mock"
	mocks2 "github.com/consensys/orchestrate/src/tx-listener/store/mocks"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/mocks"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var errMsgExceedTime = "exceeded waiting time for stopping"
var extendedWaitingTime = time.Millisecond * 500
var defaultBlockTime = time.Second * 2

func TestChainListenerSession_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apiClient := mock.NewMockOrchestrateClient(ctrl)
	ec := mock2.NewMockClient(ctrl)
	chainBlockTxsUC := mocks.NewMockChainBlock(ctrl)
	pendingJobState := mocks2.NewMockPendingJob(ctrl)
	logger := log.NewLogger()

	t.Run("should start and exit gracefully if not pending jobs after n blocks", func(t *testing.T) {
		chain := testdata.FakeChain()
		chain.ListenerBlockTimeDuration = defaultBlockTime
		tx := testdata.FakeETHTransaction()
		txHash := tx.ToETHTransaction(chain.ChainID).Hash()
		blockNumber := rand.Uint64()
		block := testdata2.FakeBlock(chain.ChainID.Uint64(), blockNumber, tx)
		cStopErr := make(chan error, 1)

		usecase := NewChainListenerSession(apiClient, ec, chainBlockTxsUC, chain, pendingJobState, logger)
		go func() {
			err := usecase.Start(ctx)
			cStopErr <- err
		}()

		proxyURL := "http://api/"+chain.UUID
		apiClient.EXPECT().ChainProxyURL(chain.UUID).Return(proxyURL)
		pendingJobState.EXPECT().ListPerChainUUID(gomock.Any(), chain.UUID).Times(waitForNEmptyBlocks+1).Return([]*entities.Job{}, nil)
		ec.EXPECT().LatestBlock(gomock.Any(), proxyURL, true).AnyTimes().Return(block, nil)
		chainBlockTxsUC.EXPECT().Execute(gomock.Any(), chain.UUID, blockNumber, []*ethcommon.Hash{&txHash}).Return(nil)

		time.Sleep(chain.ListenerBlockTimeDuration*(waitForNEmptyBlocks+1) + extendedWaitingTime)

		select {
		case <-time.Tick(extendedWaitingTime):
			t.Error(errMsgExceedTime)
		case err := <-cStopErr:
			assert.NoError(t, err)
		}
	})

	t.Run("should fail and exit gracefully if running use case fails", func(t *testing.T) {
		chain := testdata.FakeChain()
		chain.ListenerBlockTimeDuration = defaultBlockTime
		tx := testdata.FakeETHTransaction()
		txHash := tx.ToETHTransaction(chain.ChainID).Hash()
		blockNumber := rand.Uint64()
		block := testdata2.FakeBlock(chain.ChainID.Uint64(), blockNumber, tx)
		cStopErr := make(chan error, 1)

		expectedErr := fmt.Errorf("fail to run UseCase")
		usecase := NewChainListenerSession(apiClient, ec, chainBlockTxsUC, chain, pendingJobState, logger)
		go func() {
			err := usecase.Start(ctx)
			cStopErr <- err
		}()

		proxyURL := "http://api/"+chain.UUID
		apiClient.EXPECT().ChainProxyURL(chain.UUID).Return(proxyURL)
		ec.EXPECT().LatestBlock(gomock.Any(), proxyURL, true).Return(block, nil)
		chainBlockTxsUC.EXPECT().Execute(gomock.Any(), chain.UUID, blockNumber, []*ethcommon.Hash{&txHash}).Return(expectedErr)

		time.Sleep(chain.ListenerBlockTimeDuration + extendedWaitingTime)

		select {
		case <-time.Tick(extendedWaitingTime):
			t.Error(errMsgExceedTime)
		case err := <-cStopErr:
			assert.Error(t, err)
		}
	})
}