package txlistener

import (
	"context"

	backoff2 "github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/infra/messenger"
	"github.com/consensys/orchestrate/src/tx-listener/service"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/builder"
	"github.com/prometheus/client_golang/prometheus"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
)

type Service struct {
	cfg      *Config
	consumer messenger.Consumer
	logger   *log.Logger
}

func NewTxListener(cfg *Config,
	apiClient orchestrateclient.OrchestrateClient,
	ethClient ethclient.MultiClient,
	listenerMetrics prometheus.Collector,
) (*app.App, error) {
	logger := log.NewLogger()

	state := builder.NewStoreState()
	contractUCs := builder.NewContractUseCases(apiClient, ethClient, state, logger)
	jobUCs := builder.NewJobUseCases(apiClient, ethClient, contractUCs, state, logger)
	chainUCs := builder.NewChainUseCases(jobUCs, state, logger)
	sessionMngrs := builder.NewSessionManagers(apiClient, ethClient, jobUCs, chainUCs, state, logger)

	// Create service layer consumer
	bckOff := backoff2.NewConstantBackOff(cfg.RetryInterval) // @TODO Replace by config
	msgConsumer, err := service.NewMessageConsumer(cfg.Kafka, []string{cfg.ConsumerTopic},
		jobUCs.PendingJobUseCase(), jobUCs.FailedJobUseCase(), sessionMngrs.ChainSessionManager(), sessionMngrs.RetryJobSessionManager(), bckOff)
	if err != nil {
		return nil, err
	}

	txListenerSrv := &Service{
		cfg:      cfg,
		consumer: msgConsumer,
		logger:   logger,
	}

	appli, err := app.New(cfg.App, readinessOpt(apiClient, msgConsumer), app.MetricsOpt(listenerMetrics))
	if err != nil {
		return nil, err
	}

	appli.RegisterDaemon(txListenerSrv)

	return appli, nil
}

func (s *Service) Run(ctx context.Context) error {
	s.logger.Debug("starting tx-consumer service...")
	if s.cfg.IsMultiTenancyEnabled {
		ctx = multitenancy.WithUserInfo(
			authutils.WithAPIKey(ctx, s.cfg.HTTPClient.XAPIKey),
			multitenancy.NewInternalAdminUser())
	}

	// @TODO Support multiple consumer as it is in tx-sender
	err := s.consumer.Consume(ctx)
	if err != nil {
		// @TODO Exit gracefully in case of context canceled
		s.logger.WithError(err).Error("service exited with errors")
	}

	return err
}

func (s *Service) Close() error {
	return s.consumer.Close()
}

func readinessOpt(client orchestrateclient.OrchestrateClient, kafkaConsumer messenger.Consumer) app.Option {
	return func(ap *app.App) error {
		ap.AddReadinessCheck("api", client.Checker())
		ap.AddReadinessCheck("kafka", kafkaConsumer.Checker)
		return nil
	}
}
