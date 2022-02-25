package chainlistener

import (
	"context"
	"time"

	listener "github.com/consensys/orchestrate/src/chain-listener/listener/data-pullers"
	in_memory "github.com/consensys/orchestrate/src/chain-listener/state/in-memory"
	"github.com/consensys/orchestrate/src/chain-listener/use-cases/events"
	tx_listener "github.com/consensys/orchestrate/src/chain-listener/use-cases/tx-listener"
	tx_sentry "github.com/consensys/orchestrate/src/chain-listener/use-cases/tx-sentry"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/utils"
	listenermetrics "github.com/consensys/orchestrate/src/chain-listener/metrics"
	pkgsarama "github.com/consensys/orchestrate/src/infra/broker/sarama"
	"github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	"github.com/spf13/viper"
)

type Service struct {
	app *app.App
	cfg *Config
}

func NewService(ctx context.Context, cfg *Config) (*Service, error) {
	// @TODO Refactor as part of https://github.com/ConsenSys/orchestrate/issues/637
	utils.InParallel(
		func() {
			viper.Set(utils.RetryMaxIntervalViperKey, 30*time.Second)
			viper.Set(utils.RetryMaxElapsedTimeViperKey, 1*time.Hour)
			rpc.Init(ctx)
		},
	)
	logger := log.NewLogger()

	apiClient := orchestrateclient.NewHTTPClient(http.NewClient(cfg.HTTPClient), cfg.API)
	pkgsarama.InitSyncProducer(ctx)

	var listenerMetrics listenermetrics.ListenerMetrics
	if cfg.App.Metrics.IsActive(listenermetrics.ModuleName) {
		listenerMetrics = listenermetrics.NewListenerMetrics()
	} else {
		listenerMetrics = listenermetrics.NewListenerNopMetrics()
	}

	appli, err := app.New(cfg.App, ReadinessOpt(apiClient), app.MetricsOpt(listenerMetrics))
	if err != nil {
		return nil, err
	}

	ethClient := rpc.GlobalClient()

	chainInMemory := in_memory.NewChainInMemory()
	pendingJobInMemory := in_memory.NewPendingJobInMemory()
	retrySessionInMemory := in_memory.NewRetrySessionInMemory()

	sendNotificationUC := tx_listener.SendNotificationUseCase(pkgsarama.GlobalSyncProducer(),
		cfg.ChainListenerConfig.DecodedOutTopic, logger)
	updateJobStatusUC := tx_listener.UpdateJobStatusUseCase(apiClient, ethClient, sendNotificationUC, logger)
	updateChainHeadUC := tx_listener.UpdateChainHeadUseCase(apiClient, logger)
	retrySessionJobUC := tx_sentry.RetrySessionJobUseCase(apiClient, logger)
	retrySessionHandler := tx_sentry.RetrySessionHandler(apiClient, retrySessionJobUC, retrySessionInMemory, logger)

	addChainEventHandler := events.AddChainEventHandler(retrySessionHandler, chainInMemory, logger)
	updateChainEventHandler := events.UpdateChainEventHandler(retrySessionHandler, chainInMemory, logger)
	deleteChainEventHandler := events.DeleteChainEventHandler(retrySessionHandler, chainInMemory, pendingJobInMemory,
		retrySessionInMemory, logger)
	chainBlockEventHandler := events.ChainBlockEventHandler(updateJobStatusUC, retrySessionHandler, pendingJobInMemory,
		retrySessionInMemory, logger)
	pendingJobEventHandler := events.PendingJobEventHandler(apiClient, ethClient, updateJobStatusUC, retrySessionHandler,
		pendingJobInMemory, logger)

	chainListener := listener.ChainListenerService(apiClient, ethClient, addChainEventHandler, updateChainEventHandler,
		deleteChainEventHandler, cfg.ChainListenerConfig.RefreshInterval, logger)
	chainBlocksListener := listener.ChainBlockListener(apiClient, chainListener, ethClient,
		chainBlockEventHandler, updateChainHeadUC, logger)
	chainPendingJobListener := listener.ChainPendingJobsListener(apiClient, chainListener, pendingJobEventHandler,
		logger)

	appli.RegisterDaemon(chainListener)
	appli.RegisterDaemon(chainBlocksListener)
	appli.RegisterDaemon(chainPendingJobListener)
	return &Service{
		app: appli,
		cfg: cfg,
	}, nil
}

func (s *Service) Run(ctx context.Context) error {
	var err error
	if s.cfg.IsMultiTenancyEnabled {
		ctx = multitenancy.WithUserInfo(
			authutils.WithAPIKey(ctx, s.cfg.HTTPClient.XAPIKey),
			multitenancy.NewInternalAdminUser())
	}

	err = s.app.Run(ctx)
	return err
}

func ReadinessOpt(client orchestrateclient.OrchestrateClient) app.Option {
	return func(ap *app.App) error {
		ap.AddReadinessCheck("api", client.Checker())
		ap.AddReadinessCheck("kafka", pkgsarama.GlobalClientChecker())
		return nil
	}
}
