package datapullers

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/chain-listener/service/listener"
	"github.com/consensys/orchestrate/src/chain-listener/service/listener/events"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	ethclientutils "github.com/consensys/orchestrate/src/infra/ethclient/utils"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const listenBlocksComponent = "chain-listener.data-puller.chain-blocks"
const listenBlocksSessionComponent = "chain-listener.data-puller.chain-blocks.session"

type listenBlockListener struct {
	ethClient         ethclient.Client
	apiClient         orchestrateclient.OrchestrateClient
	onChainBlockEvent usecases.ChainBlockTxsUseCase
	chainListener     listener.ChainListener
	sessions          map[string]*ListenBlockSession
	logger            *log.Logger
	cancelCtx         context.CancelFunc
	cerr              chan error
}

func ChainBlockListener(apiClient orchestrateclient.OrchestrateClient,
	chainListener listener.ChainListener,
	ethClient ethclient.Client,
	onChainBlockEvent usecases.ChainBlockTxsUseCase,
	logger *log.Logger,
) listener.ChainBlockListener {
	return &listenBlockListener{
		ethClient:         ethClient,
		apiClient:         apiClient,
		chainListener:     chainListener,
		onChainBlockEvent: onChainBlockEvent,
		sessions:          make(map[string]*ListenBlockSession),
		logger:            logger.SetComponent(listenBlocksComponent),
		cerr:              make(chan error, 1),
	}
}

func (l *listenBlockListener) Run(ctx context.Context) error {
	l.logger.Info("chain block session builder started")
	ctx, l.cancelCtx = context.WithCancel(ctx)
	go func() {
		chainEvents := make(chan *events.Chain, 10) // @TODO Optimize chan size
		defer close(chainEvents)
		subscriptionID := l.chainListener.Subscribe(chainEvents)
		for event := range chainEvents {
			switch event.Type {
			case events.NewChain:
				err := l.addChainSession(ctx, event.Chain)
				if err != nil && !errors.IsAlreadyExistsError(err) {
					l.cerr <- err
					break
				}
			case events.UpdatedChain:
				err := l.updateChainSession(ctx, event.Chain)
				if err != nil {
					l.cerr <- err
					break
				}
			case events.DeletedChain:
				err := l.removeChainSession(ctx, event.Chain)
				if err != nil {
					l.cerr <- err
					break
				}
			}
		}

		_ = l.chainListener.Unsubscribe(subscriptionID)
	}()

	select {
	case err := <-l.cerr:
		l.logger.WithField("err", err).Info("chain block session builder exited with errors")
		return err
	case <-ctx.Done():
		l.logger.WithField("reason", ctx.Err().Error()).Info("chain block session builder gracefully stopping...")
		return nil
	}
}

func (l *listenBlockListener) Close() error {
	l.cancelCtx()
	gerr := []error{}
	for _, sess := range l.sessions {
		if err := sess.Close(); err != nil {
			gerr = append(gerr, err)
		}
	}

	// @TODO Concatenate errors
	if len(gerr) > 1 {
		return gerr[0]
	}

	return nil
}

func (l *listenBlockListener) addChainSession(ctx context.Context, chain *entities.Chain) error {
	sessID := chain.UUID
	if _, ok := l.sessions[sessID]; ok {
		errMsg := "chain block session already exist"
		l.logger.WithField("chain", chain.UUID).Error(errMsg)
		return errors.AlreadyExistsError(errMsg)
	}

	sess := NewListenBlockSession(l.apiClient, l.ethClient, l.onChainBlockEvent, chain, l.logger)
	l.sessions[sessID] = sess
	go func(s *ListenBlockSession) {
		err := s.Run(ctx)
		if err != nil {
			l.cerr <- err
		}
	}(sess)
	return nil
}

func (l *listenBlockListener) removeChainSession(_ context.Context, chain *entities.Chain) error {
	l.logger.Debug("chain block removing session...")
	sessID := chain.UUID
	if sess, ok := l.sessions[sessID]; ok {
		return sess.Close()
	}

	errMsg := "chain block session does not exist for removing"
	l.logger.WithField("chain", chain.UUID).Error(errMsg)
	return errors.NotFoundError(errMsg)
}

func (l *listenBlockListener) updateChainSession(_ context.Context, chain *entities.Chain) error {
	l.logger.Debug("updating chain block session...")
	sessID := chain.UUID
	if sess, ok := l.sessions[sessID]; ok {
		return sess.Update(chain)
	}

	errMsg := "chain block session does not exist for updating"
	l.logger.WithField("chain", chain.UUID).Error(errMsg)
	return errors.NotFoundError(errMsg)
}

type ListenBlockSession struct {
	ethClient       ethclient.Client
	apiClient       orchestrateclient.OrchestrateClient
	chainBlockTxsUC usecases.ChainBlockTxsUseCase
	logger          *log.Logger
	chain           *entities.Chain
	cancelCtx       context.CancelFunc
	curBlockNumber  uint64
	interval        time.Duration
	ticker          *time.Ticker
	cerr            chan error
}

func NewListenBlockSession(apiClient orchestrateclient.OrchestrateClient,
	ethClient ethclient.Client,
	chainBlockTxsUC usecases.ChainBlockTxsUseCase,
	chain *entities.Chain,
	logger *log.Logger,
) *ListenBlockSession {
	return &ListenBlockSession{
		ethClient:       ethClient,
		apiClient:       apiClient,
		chain:           chain,
		chainBlockTxsUC: chainBlockTxsUC,
		logger:          logger.WithField("chain", chain.UUID).SetComponent(listenBlocksSessionComponent),
		curBlockNumber:  chain.ListenerCurrentBlock,
		interval:        chain.ListenerBackOffDuration,
		cerr:            make(chan error, 1),
	}
}

func (s *ListenBlockSession) Run(ctx context.Context) error {
	s.logger.WithField("refresh", s.interval.String()).Info("chain block listener started")
	ctx, s.cancelCtx = context.WithCancel(ctx)

	err := s.runIt(ctx, s.curBlockNumber)
	if err != nil {
		return err
	}
	go func() {
		s.ticker = time.NewTicker(s.interval)
		defer s.ticker.Stop()
		for {
			<-s.ticker.C
			err = s.runIt(ctx, s.curBlockNumber)
			if err != nil {
				if ctx.Err() == nil { // Context is not done
					s.cerr <- err
				}
				return
			}
		}
	}()

	select {
	case err := <-s.cerr:
		s.logger.WithField("err", err).Info("chain block session exited with errors")
		return err
	case <-ctx.Done():
		s.logger.WithField("reason", ctx.Err().Error()).Info("chain block session gracefully stopping...")
		return nil
	}
}

func (s *ListenBlockSession) Close() error {
	defer close(s.cerr)
	s.cancelCtx()
	s.logger.Info("chain block session listener closed")
	return nil
}

func (s *ListenBlockSession) Update(chain *entities.Chain) error {
	s.logger.Info("chain block session listener updated")
	if s.chain.ListenerBackOffDuration.Milliseconds() != chain.ListenerBackOffDuration.Milliseconds() {
		s.ticker.Reset(chain.ListenerBackOffDuration)
	}

	s.chain = chain
	return nil
}

func (s *ListenBlockSession) runIt(ctx context.Context, curBlockNumber uint64) error {
	blockEvents, nextBlockNumber, err := s.retrieveBlocks(ctx, curBlockNumber)
	if err != nil {
		return err
	}

	err = s.triggerEvents(ctx, blockEvents)
	if err != nil {
		return err
	}

	s.curBlockNumber = nextBlockNumber
	return nil
}

func (s *ListenBlockSession) triggerEvents(ctx context.Context, blockEvents []*events.Block) error {
	// @TODO Can I run it in parallel ???
	for _, event := range blockEvents {
		if err := s.chainBlockTxsUC.Execute(ctx, s.chain.UUID, event.Number, event.TxHashes); err != nil {
			return err
		}
	}

	return nil
}

func (s *ListenBlockSession) retrieveBlocks(ctx context.Context, curBlockNumber uint64) ([]*events.Block, uint64, error) {
	chainURL := s.apiClient.ChainProxyURL(s.chain.UUID)
	newBlockEvents := []*events.Block{}

	latestBlock, err := s.retrieveBlock(ctx, chainURL, "latest")
	if err != nil {
		return nil, curBlockNumber, err
	}

	if latestBlock.NumberU64() <= curBlockNumber {
		return newBlockEvents, curBlockNumber, nil
	}

	wg := &sync.WaitGroup{}

	// @TODO Set a max number of parallel calls
	for fromBlock := curBlockNumber + 1; fromBlock < latestBlock.NumberU64(); fromBlock++ {
		wg.Add(1)
		go func(blockNumber uint64) {
			defer wg.Done()
			block, err := s.retrieveBlock(ctx, chainURL, blockNumber)
			if err != nil {
				s.cerr <- err
				return
			}
			newBlockEvents = append(newBlockEvents, events.NewEthereumBlock(s.chain.UUID, block))
		}(fromBlock)
	}

	wg.Wait()
	newBlockEvents = append(newBlockEvents, events.NewEthereumBlock(s.chain.UUID, latestBlock))
	return newBlockEvents, latestBlock.NumberU64(), nil
}

func (s *ListenBlockSession) retrieveBlock(ctx context.Context, chainURL string, blockNumber interface{}) (*ethtypes.Block, error) {
	var chainBlock *ethtypes.Block
	var err error

	// @TODO if includeTxs==false, it fails because Transactions struct does not match expected
	cctx := ethclientutils.RetryConnectionError(ctx, true)
	if bn, ok := blockNumber.(uint64); ok {
		chainBlock, err = s.ethClient.BlockByNumber(cctx, chainURL, new(big.Int).SetUint64(bn), true)
	} else {
		chainBlock, err = s.ethClient.LatestBlock(cctx, chainURL, true)
	}

	if err != nil {
		errMsg := "failed to fetch blocks from chain"
		s.logger.WithError(err).Error(errMsg)
		return nil, errors.DependencyFailureError(errMsg)
	}

	s.logger.WithField("block", blockNumber).Debug("retrieved block")
	return chainBlock, nil
}
