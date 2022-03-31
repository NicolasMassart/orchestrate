package txlistener

import (
	"context"

	"github.com/consensys/orchestrate/src/infra/ethclient"
	listener "github.com/consensys/orchestrate/src/tx-listener/service/listener/data-pullers"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/builder"
	"github.com/prometheus/client_golang/prometheus"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
)

type Service struct {
	App *app.App
	cfg *Config
}

func NewTxlistener(cfg *Config,
	apiClient orchestrateclient.OrchestrateClient,
	ethClient ethclient.MultiClient,
	listenerMetrics prometheus.Collector,
) (*Service, error) {
	appli, err := app.New(cfg.App, ReadinessOpt(apiClient), app.MetricsOpt(listenerMetrics))
	if err != nil {
		return nil, err
	}

	logger := log.NewLogger()

	ucs := builder.NewEventUseCases(apiClient, ethClient, logger)

	txlistener := listener.TxlistenerService(apiClient, ethClient, ucs.AddChainUseCase(), ucs.UpdateChainUseCase(),
		ucs.DeleteChainUseCase(), cfg.RefreshInterval, logger)
	chainBlocksListener := listener.ChainBlockListener(apiClient, txlistener, ethClient,
		ucs.ChainBlockTxsUseCase(), logger)
	chainPendingJobListener := listener.ChainPendingJobsListener(apiClient, txlistener, ucs.PendingJobUseCase(),
		cfg.RefreshInterval, logger)

	appli.RegisterDaemon(txlistener)
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
		return nil
	}
}
