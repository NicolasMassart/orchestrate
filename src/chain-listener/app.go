package chainlistener

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases/events"
	tx_listener "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases/tx-listener"
	tx_sentry "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases/tx-sentry"
	listener "github.com/consensys/orchestrate/src/chain-listener/service/listener/data-pullers"
	in_memory "github.com/consensys/orchestrate/src/chain-listener/store/in-memory"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/prometheus/client_golang/prometheus"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	pkgsarama "github.com/consensys/orchestrate/src/infra/broker/sarama"
)

type Service struct {
	App *app.App
	cfg *Config
}

func NewChainListener(cfg *Config,
	apiClient orchestrateclient.OrchestrateClient,
	saramaCli sarama.SyncProducer,
	ethClient ethclient.MultiClient,
	listenerMetrics prometheus.Collector,
) (*Service, error) {
	appli, err := app.New(cfg.App, ReadinessOpt(apiClient), app.MetricsOpt(listenerMetrics))
	if err != nil {
		return nil, err
	}

	logger := log.NewLogger()

	chainInMemory := in_memory.NewChainInMemory()
	pendingJobInMemory := in_memory.NewPendingJobInMemory()
	retrySessionInMemory := in_memory.NewRetrySessionInMemory()

	sendNotificationUC := tx_listener.SendNotificationUseCase(apiClient, saramaCli,
		cfg.ChainListenerConfig.DecodedOutTopic, logger)
	registerDeployedContractUC := tx_listener.RegisterDeployedContractUseCase(apiClient, ethClient, chainInMemory, logger)
	updateJobStatusUC := tx_listener.NotifyMinedJobUseCase(apiClient, ethClient, sendNotificationUC,
		registerDeployedContractUC, chainInMemory, logger)
	updateChainHeadUC := tx_listener.UpdateChainHeadUseCase(apiClient, logger)
	retrySessionJobUC := tx_sentry.RetrySessionJobUseCase(apiClient, logger)
	retrySessionHandler := tx_sentry.RetrySessionManager(apiClient, retrySessionJobUC, retrySessionInMemory, logger)

	addChainEventUC := events.AddChainUseCase(retrySessionHandler, chainInMemory, logger)
	updateChainEventUC := events.UpdateChainUseCase(retrySessionHandler, chainInMemory, logger)
	deleteChainEventUC := events.DeleteChainUseCase(retrySessionHandler, chainInMemory, pendingJobInMemory,
		retrySessionInMemory, logger)
	chainBlockEventHandler := events.ChainBlockTxsUseCase(updateJobStatusUC, retrySessionHandler, pendingJobInMemory,
		retrySessionInMemory, logger)
	pendingJobEventHandler := events.PendingJobUseCase(apiClient, ethClient, updateJobStatusUC, retrySessionHandler,
		pendingJobInMemory, logger)

	chainListener := listener.ChainListenerService(apiClient, ethClient, addChainEventUC, updateChainEventUC,
		deleteChainEventUC, cfg.ChainListenerConfig.RefreshInterval, logger)
	chainBlocksListener := listener.ChainBlockListener(apiClient, chainListener, ethClient,
		chainBlockEventHandler, updateChainHeadUC, logger)
	chainPendingJobListener := listener.ChainPendingJobsListener(apiClient, chainListener, pendingJobEventHandler,
		cfg.ChainListenerConfig.RefreshInterval, logger)

	appli.RegisterDaemon(chainListener)
	appli.RegisterDaemon(chainBlocksListener)
	appli.RegisterDaemon(chainPendingJobListener)
	return &Service{
		appli, cfg,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	var err error
	if s.cfg.IsMultiTenancyEnabled {
		ctx = multitenancy.WithUserInfo(
			authutils.WithAPIKey(ctx, s.cfg.HTTPClient.XAPIKey),
			multitenancy.NewInternalAdminUser())
	}

	err = s.App.Run(ctx)
	return err
}

func (s *Service) Stop(ctx context.Context) error {
	return s.App.Stop(ctx)
}

func ReadinessOpt(client orchestrateclient.OrchestrateClient) app.Option {
	return func(ap *app.App) error {
		ap.AddReadinessCheck("api", client.Checker())
		ap.AddReadinessCheck("kafka", pkgsarama.GlobalClientChecker())
		return nil
	}
}
