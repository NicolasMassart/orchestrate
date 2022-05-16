package chains

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const listenBlocksComponent = "tx-listener.chains.session-manager"

type ChainSessionMngr struct {
	ethClient          ethclient.Client
	apiClient          sdk.OrchestrateClient
	chainUCs           usecases.ChainUseCases
	logger             *log.Logger
	pendingJobState    store.PendingJob
	subscriptionsState store.Subscriptions
	chainState         store.Chain
}

func ChainSessionManager(apiClient sdk.OrchestrateClient,
	ethClient ethclient.Client,
	chainUCs usecases.ChainUseCases,
	pendingJobState store.PendingJob,
	subscriptionsState store.Subscriptions,
	chainState store.Chain,
	logger *log.Logger,
) *ChainSessionMngr {
	return &ChainSessionMngr{
		ethClient:          ethClient,
		apiClient:          apiClient,
		chainUCs:           chainUCs,
		pendingJobState:    pendingJobState,
		subscriptionsState: subscriptionsState,
		chainState:         chainState,
		logger:             logger.SetComponent(listenBlocksComponent),
	}
}

func (l *ChainSessionMngr) StartSession(ctx context.Context, chainUUID string) error {
	logger := l.logger.WithField("chain", chainUUID)

	chain, err := l.chainState.Get(ctx, chainUUID)
	if err != nil && !errors.IsNotFoundError(err) {
		return err
	}

	// It exits in case of active session
	if chain != nil {
		// @TODO Evaluate consequence of chain update
		errMsg := "chain listening session already exists"
		logger.Debug(errMsg)
		return errors.AlreadyExistsError(errMsg)
	}

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

	sess := NewChainListenerSession(l.apiClient, l.ethClient, l.chainUCs.ChainBlockTxsUseCase(), l.chainUCs.ChainBlockEventsUseCase(),
		chain, l.pendingJobState, l.subscriptionsState, l.logger)

	go func(s *ChainListenerSession) {
		err = backoff.RetryNotify(func() error {
			err2 := s.Start(ctx)
			switch {
			case err2 == nil:
				return nil
			case err2 == context.DeadlineExceeded || err2 == context.Canceled:
				return nil
			case ctx.Err() != nil:
				return nil
			default:
				return err
			}
		}, backoff.NewConstantBackOff(time.Second*2), func(e error, duration time.Duration) {
			log.FromContext(ctx).
				WithError(e).
				Warnf("failed to run chain listener, restarting session in %v...", duration)
		})

		if err != nil {
			logger.WithError(err).Error("chain listener session has exited with errors")
			return
		}

		err = l.removeSession(ctx, chainUUID)
		if err != nil {
			errMsg := "failed to removing chain listener session"
			logger.WithError(err).Error(errMsg)
		}

		logger.Info("chain listener session has exited gracefully")
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
