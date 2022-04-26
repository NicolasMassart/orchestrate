package chains

import (
	"context"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/sessions"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const listenBlocksComponent = "tx-listener.chains.session-manager"

type ChainSessionMngr struct {
	ethClient           ethclient.Client
	apiClient           orchestrateclient.OrchestrateClient
	chainBlockTxsUC     usecases.ChainBlock
	retrySessionManager sessions.TxSentrySessionManager
	logger              *log.Logger
	pendingJobState     store.PendingJob
	retrySessionState   store.RetrySessions
	chainState          store.Chain
}

func ChainSessionManager(apiClient orchestrateclient.OrchestrateClient,
	ethClient ethclient.Client,
	retrySessionManager sessions.TxSentrySessionManager,
	chainBlockTxsUC usecases.ChainBlock,
	pendingJobState store.PendingJob,
	chainState store.Chain,
	retrySessionState store.RetrySessions,
	logger *log.Logger,
) *ChainSessionMngr {
	return &ChainSessionMngr{
		ethClient:           ethClient,
		apiClient:           apiClient,
		retrySessionManager: retrySessionManager,
		chainBlockTxsUC:     chainBlockTxsUC,
		pendingJobState:     pendingJobState,
		retrySessionState:   retrySessionState,
		chainState:          chainState,
		logger:              logger.SetComponent(listenBlocksComponent),
	}
}

func (l *ChainSessionMngr) StartSession(ctx context.Context, chainUUID string) error {
	logger := l.logger.WithField("chain", chainUUID)

	chain, err := l.chainState.Get(ctx, chainUUID)
	if err != nil {
		return err
	}

	// It exits in case of active session
	if chain != nil {
		logger.Debug("chain listening session already exists")
		return nil
	}

	// @TODO Evaluate consequence of chain update
	chainResp, err := l.apiClient.GetChain(ctx, chainUUID)
	if err != nil {
		errMsg := "failed to fetch chain data"
		logger.WithError(err).Error(errMsg)
		return err
	}

	chain = formatters.ChainResponseToEntity(chainResp)
	err = l.chainState.Add(ctx, chain)
	if err != nil {
		return err
	}

	sess := NewChainListenerSession(l.apiClient, l.ethClient, l.chainBlockTxsUC, chain, l.pendingJobState, l.logger)

	go func(s *ChainListenerSession) {
		// @TODO How to propagate error ???
		err := s.Start(ctx)
		if err != nil {
			errMsg := "failed to run retry session"
			logger.WithError(err).Error(errMsg)
		}

		err = l.removeSession(ctx, chainUUID)
		if err != nil {
			errMsg := "failed to removing session"
			logger.WithError(err).Error(errMsg)
		}
	}(sess)

	return nil
}

func (l *ChainSessionMngr) removeSession(ctx context.Context, chainUUID string) error {
	logger := l.logger.WithField("chain", chainUUID)
	err := l.chainState.Delete(ctx, chainUUID)
	if err != nil {
		errMsg := "failed to delete chain state"
		logger.WithError(err).Error(errMsg)
		return err
	}

	logger.Debug("session has been removed")
	return nil
}
