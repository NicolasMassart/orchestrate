package txlistener

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/infra/messenger"
	"github.com/consensys/orchestrate/src/tx-listener/service"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/builder"
	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
)

type Service struct {
	cfg       *Config
	consumers []messenger.Consumer
	logger    *log.Logger
	cancel    context.CancelFunc
}

func NewTxListener(cfg *Config,
	apiClient sdk.OrchestrateClient,
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
	bckOff := backoff.NewConstantBackOff(cfg.RetryInterval) // @TODO Replace by config
	consumers := make([]messenger.Consumer, cfg.Kafka.NConsumers)
	for idx := 0; idx < cfg.Kafka.NConsumers; idx++ {
		var err error
		consumers[idx], err = service.NewMessageConsumer(cfg.Kafka, []string{cfg.ConsumerTopic},
			jobUCs.PendingJobUseCase(), jobUCs.FailedJobUseCase(), sessionMngrs.ChainSessionManager(), sessionMngrs.RetryJobSessionManager(), bckOff)
		if err != nil {
			return nil, err
		}
	}

	txListenerSrv := &Service{
		cfg:       cfg,
		consumers: consumers,
		logger:    logger,
	}

	appli, err := app.New(cfg.App, readinessOpt(apiClient, consumers[0]), app.MetricsOpt(listenerMetrics))
	if err != nil {
		return nil, err
	}

	appli.RegisterDaemon(txListenerSrv)

	return appli, nil
}

func (s *Service) Run(ctx context.Context) error {
	s.logger.Debug("starting tx-listener service")
	if s.cfg.IsMultiTenancyEnabled {
		ctx = multitenancy.WithUserInfo(
			authutils.WithAPIKey(ctx, s.cfg.HTTPClient.XAPIKey),
			multitenancy.NewInternalAdminUser())
	}

	ctx, s.cancel = context.WithCancel(ctx)
	gr := &multierror.Group{}
	for idx, consumerGroup := range s.consumers {
		cGroup := consumerGroup
		cGroupID := fmt.Sprintf("c-%d", idx)
		logger := s.logger.WithField("consumer", cGroupID)
		cctx := log.With(log.WithField(ctx, "consumer", cGroupID), logger)
		gr.Go(func() error {
			// We retry once after consume exits to prevent entire stack to exit after kafka rebalance is triggered
			err := backoff.RetryNotify(
				func() error {
					err := cGroup.Consume(cctx)

					// In this case, kafka rebalance was triggered and we want to retry
					if err == nil && cctx.Err() == nil {
						return fmt.Errorf("kafka rebalance was triggered")
					}

					return backoff.Permanent(err)
				},
				backoff.NewConstantBackOff(time.Millisecond*500),
				func(err error, duration time.Duration) {
					logger.WithError(err).Warnf("consuming session exited, retrying in %s", duration.String())
				},
			)
			s.cancel()
			return err
		})
	}

	return gr.Wait().ErrorOrNil()
}

func (s *Service) Close() error {
	var gerr error
	for _, consumerGroup := range s.consumers {
		gerr = errors.CombineErrors(gerr, consumerGroup.Close())
	}

	return gerr
}

func readinessOpt(client sdk.OrchestrateClient, kafkaConsumer messenger.Consumer) app.Option {
	return func(ap *app.App) error {
		ap.AddReadinessCheck("api", client.Checker())
		ap.AddReadinessCheck("kafka", kafkaConsumer.Checker)
		return nil
	}
}
