package chainlistener

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/src/chain-listener/chain-listener/builder"
	listener "github.com/consensys/orchestrate/src/chain-listener/service/listener/data-pullers"
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

	ucs := builder.NewEventUseCases(apiClient, saramaCli, ethClient, cfg.ChainListenerConfig.DecodedOutTopic, logger)

	chainListener := listener.ChainListenerService(apiClient, ethClient, ucs.AddChainUseCase(), ucs.UpdateChainUseCase(),
		ucs.DeleteChainUseCase(), cfg.ChainListenerConfig.RefreshInterval, logger)
	chainBlocksListener := listener.ChainBlockListener(apiClient, chainListener, ethClient,
		ucs.ChainBlockTxsUseCase(), logger)
	chainPendingJobListener := listener.ChainPendingJobsListener(apiClient, chainListener, ucs.PendingJobUseCase(),
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
