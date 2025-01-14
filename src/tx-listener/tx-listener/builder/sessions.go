package builder

import (
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/sessions"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/sessions/chains"
	tx_sentry "github.com/consensys/orchestrate/src/tx-listener/tx-listener/sessions/tx-sentry"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

type sessionMngrs struct {
	chainSessionMngr    sessions.ChainSessionManager
	retryJobSessionMngr sessions.RetryJobSessionManager
}

func (b *sessionMngrs) ChainSessionManager() sessions.ChainSessionManager {
	return b.chainSessionMngr
}

func (b *sessionMngrs) RetryJobSessionManager() sessions.RetryJobSessionManager {
	return b.retryJobSessionMngr
}

func NewSessionManagers(messengerAPI sdk.MessengerAPI,
	apiClient sdk.OrchestrateClient,
	ethClient ethclient.MultiClient,
	jobUCs usecases.JobUseCases,
	chainUCs usecases.ChainUseCases,
	state store.State,
	logger *log.Logger,
) sessions.SessionManagers {
	retryJobSessionMngr := tx_sentry.NewRetrySessionManager(messengerAPI, apiClient, jobUCs.RetryJobUseCase(), state.RetryJobSessionState(),
		state.PendingJobState(), logger)
	chainSessionMngr := chains.ChainSessionManager(apiClient, ethClient, chainUCs, state.PendingJobState(),
		state.SubscriptionState(), state.ChainState(), logger)

	return &sessionMngrs{
		chainSessionMngr:    chainSessionMngr,
		retryJobSessionMngr: retryJobSessionMngr,
	}
}
