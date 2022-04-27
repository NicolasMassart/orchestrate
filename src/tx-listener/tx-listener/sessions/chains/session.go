package chains

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	ethclientutils "github.com/consensys/orchestrate/src/infra/ethclient/utils"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const listenBlocksSessionComponent = "tx-listener.chains.session"
const waitForNEmptyBlocks = 3

type ChainListenerSession struct {
	ethClient         ethclient.Client
	apiClient         orchestrateclient.OrchestrateClient
	chainBlockTxsUC   usecases.ChainBlock
	logger            *log.Logger
	chain             *entities.Chain
	cancelCtx         context.CancelFunc
	curBlockNumber    uint64
	pendingJobState   store.PendingJob
	blockTimeDuration time.Duration
	ticker            *time.Ticker
	cerr              chan error
}

func NewChainListenerSession(apiClient orchestrateclient.OrchestrateClient,
	ethClient ethclient.Client,
	chainBlockTxsUC usecases.ChainBlock,
	chain *entities.Chain,
	pendingJobState store.PendingJob,
	logger *log.Logger,
) *ChainListenerSession {
	return &ChainListenerSession{
		ethClient:         ethClient,
		apiClient:         apiClient,
		chain:             chain,
		pendingJobState:   pendingJobState,
		chainBlockTxsUC:   chainBlockTxsUC,
		logger:            logger.WithField("chain", chain.UUID).SetComponent(listenBlocksSessionComponent),
		blockTimeDuration: chain.ListenerBlockTimeDuration,
		cerr:              make(chan error, 1),
	}
}

func (s *ChainListenerSession) Start(ctx context.Context) error {
	s.logger.WithField("block_time", s.blockTimeDuration.String()).Info("chain block listener started")
	ctx, s.cancelCtx = context.WithCancel(ctx)

	proxyChainURL := s.apiClient.ChainProxyURL(s.chain.UUID)
	err := s.runIt(ctx, s.curBlockNumber, proxyChainURL)
	if err != nil {
		return err
	}

	go func() {
		nBlockWithoutPendingJobs := 0
		s.ticker = time.NewTicker(s.blockTimeDuration)
		defer s.ticker.Stop()
		for {
			<-s.ticker.C
			err = s.runIt(ctx, s.curBlockNumber, proxyChainURL)
			if err != nil {
				if ctx.Err() == nil { // Context is not done
					s.cerr <- err
				}
				return
			}

			pendingJobs, err := s.pendingJobState.ListPerChainUUID(ctx, s.chain.UUID)
			if err != nil {
				if ctx.Err() == nil { // Context is not done
					s.cerr <- err
				}
				return
			}

			if len(pendingJobs) == 0 {
				if nBlockWithoutPendingJobs >= waitForNEmptyBlocks {
					s.logger.Debug("no pending jobs. Stopping session...")
					s.Stop()
					return
				}

				nBlockWithoutPendingJobs++
			} else {
				nBlockWithoutPendingJobs = 0
			}
		}
	}()

	select {
	case err := <-s.cerr:
		s.logger.WithField("err", err).Info("chain block session exited with errors")
		return err
	case <-ctx.Done():
		s.logger.WithField("reason", ctx.Err().Error()).Info("chain block session gracefully stopped")
		return nil
	}
}

func (s *ChainListenerSession) Stop() {
	s.cancelCtx()
	s.logger.Debug("chain block session listener closed")
}

func (s *ChainListenerSession) runIt(ctx context.Context, curBlockNumber uint64, proxyChainURL string) error {
	blockEvents, nextBlockNumber, err := s.retrieveBlocks(ctx, curBlockNumber, proxyChainURL)
	if err != nil {
		return err
	}

	// @TODO Can I run it in parallel ???
	for _, event := range blockEvents {
		if err := s.chainBlockTxsUC.Execute(ctx, s.chain.UUID, event.Number, event.TxHashes); err != nil {
			return err
		}
	}

	s.curBlockNumber = nextBlockNumber
	return nil
}

func (s *ChainListenerSession) retrieveBlocks(ctx context.Context, curBlockNumber uint64, proxyChainURL string) ([]*Block, uint64, error) {
	newBlockEvents := []*Block{}

	latestBlock, err := s.retrieveBlock(ctx, proxyChainURL, "latest")
	if err != nil {
		return nil, curBlockNumber, err
	}

	if latestBlock == nil || latestBlock.NumberU64() <= curBlockNumber {
		return newBlockEvents, curBlockNumber, nil
	}

	if curBlockNumber != 0 {
		wg := &sync.WaitGroup{}
		// @TODO Set a max number of parallel calls
		for fromBlock := curBlockNumber + 1; fromBlock < latestBlock.NumberU64(); fromBlock++ {
			wg.Add(1)
			go func(blockNumber uint64) {
				defer wg.Done()
				block, err := s.retrieveBlock(ctx, proxyChainURL, blockNumber)
				if err != nil {
					s.cerr <- err
					return
				}
				newBlockEvents = append(newBlockEvents, NewEthereumBlock(s.chain.UUID, block))
			}(fromBlock)
		}
		wg.Wait()
	}

	newBlockEvents = append(newBlockEvents, NewEthereumBlock(s.chain.UUID, latestBlock))
	return newBlockEvents, latestBlock.NumberU64(), nil
}

func (s *ChainListenerSession) retrieveBlock(ctx context.Context, chainURL string, blockNumber interface{}) (*ethtypes.Block, error) {
	var chainBlock *ethtypes.Block
	var err error

	// @TODO Evaluate consequences of retrying "true"
	// For instance, in case chain was deleted we will loop over a 404 blocking the consumer queue
	cctx := ethclientutils.RetryConnectionError(ctx, true)
	if bn, ok := blockNumber.(uint64); ok {
		chainBlock, err = s.ethClient.BlockByNumber(cctx, chainURL, new(big.Int).SetUint64(bn), true)
	} else {
		chainBlock, err = s.ethClient.LatestBlock(cctx, chainURL, true)
	}

	if err != nil {
		if ctx.Err() != nil {
			return nil, nil
		}

		errMsg := "failed to fetch blocks from chain"
		s.logger.WithError(err).Error(errMsg)
		return nil, errors.DependencyFailureError(errMsg)
	}

	s.logger.WithField("block", blockNumber).Debug("retrieved block")
	return chainBlock, nil
}
