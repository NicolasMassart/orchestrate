package builder

import (
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/chains"
)

type chainUCs struct {
	chainBlockTxsUC    usecases.ChainBlockTxs
	chainBlockEventsUC usecases.ChainBlockEvents
}

func (s *chainUCs) ChainBlockTxsUseCase() usecases.ChainBlockTxs {
	return s.chainBlockTxsUC
}

func (s *chainUCs) ChainBlockEventsUseCase() usecases.ChainBlockEvents {
	return s.chainBlockEventsUC
}

func NewChainUseCases(proxyClient sdk.ChainProxyClient,
	ethClient ethclient.Client,
	jobUCs usecases.JobUseCases,
	subscriptionUCs usecases.SubscriptionUseCases,
	state store.State,
	logger *log.Logger,
) usecases.ChainUseCases {
	chainBlockTxs := chains.NewChainBlockTxsUseCase(jobUCs.MinedJobUseCase(), 
		state.PendingJobState(), logger)
	chainBlockEvents := chains.NewChainBlockEventsUseCase(proxyClient, ethClient, subscriptionUCs.NotifySubscriptionEventsUseCase(), 
		state.SubscriptionState(), logger)
	
	return &chainUCs{
		chainBlockTxsUC: chainBlockTxs,
		chainBlockEventsUC: chainBlockEvents,
	}
}
