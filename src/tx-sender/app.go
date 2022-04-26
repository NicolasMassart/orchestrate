package txsender

import (
	"context"
	"fmt"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/infra/messenger"
	"github.com/consensys/orchestrate/src/infra/messenger/kafka"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/nonce"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/nonce/manager"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/errors"
	api "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/infra/redis"
	"github.com/consensys/orchestrate/src/tx-sender/service"
	"github.com/consensys/orchestrate/src/tx-sender/store/memory"
	redisnoncemngr "github.com/consensys/orchestrate/src/tx-sender/store/redis"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/builder"
	keymanager "github.com/consensys/quorum-key-manager/pkg/client"
	"github.com/hashicorp/go-multierror"
)

const component = "application"

type txSenderDaemon struct {
	keyManagerClient keymanager.KeyManagerClient
	jobClient        api.JobClient
	ec               ethclient.MultiClient
	nonceManager     nonce.Manager
	consumers        []messenger.Consumer
	config           *Config
	logger           *log.Logger
	cancel           context.CancelFunc
}

func NewTxSender(
	config *Config,
	keyManagerClient keymanager.KeyManagerClient,
	apiClient api.OrchestrateClient,
	ec ethclient.MultiClient,
	redisCli redis.Client,
) (*app.App, error) {
	var nm nonce.Manager
	if config.NonceManagerType == NonceManagerTypeInMemory {
		nm = manager.NewNonceManager(ec, memory.NewNonceSender(config.NonceManagerExpiration), memory.NewNonceRecoveryTracker(),
			config.ProxyURL, config.NonceMaxRecovery)
	} else if config.NonceManagerType == NonceManagerTypeRedis {
		nm = manager.NewNonceManager(ec, redisnoncemngr.NewNonceSender(redisCli, config.NonceManagerExpiration), redisnoncemngr.NewNonceRecoveryTracker(redisCli),
			config.ProxyURL, config.NonceMaxRecovery)
	}

	// Create business layer use cases
	useCases := builder.NewUseCases(apiClient, keyManagerClient, ec, nm, config.ProxyURL)
	msgConsumerHandler := service.NewMessageConsumerHandler(useCases, apiClient, config.BckOff)

	consumers := make([]messenger.Consumer, config.NConsumer)
	for idx := 0; idx < config.NConsumer; idx++ {
		var err error
		consumers[idx], err = kafka.NewMessageConsumer(config.Kafka, []string{config.ConsumerTopic}, msgConsumerHandler)
		if err != nil {
			return nil, err
		}
	}

	txSenderDaemon := &txSenderDaemon{
		keyManagerClient: keyManagerClient,
		jobClient:        apiClient,
		consumers:        consumers,
		config:           config,
		ec:               ec,
		nonceManager:     nm,
		logger:           log.NewLogger().SetComponent(component),
	}

	appli, err := app.New(config.App, readinessOpt(apiClient, redisCli, consumers[0]), app.MetricsOpt())
	if err != nil {
		return nil, err
	}

	appli.RegisterDaemon(txSenderDaemon)

	return appli, nil
}

func (d *txSenderDaemon) Run(ctx context.Context) error {
	d.logger.Debug("starting transaction sender")

	ctx, d.cancel = context.WithCancel(ctx)
	if d.config.IsMultiTenancyEnabled {
		ctx = multitenancy.WithUserInfo(ctx, multitenancy.NewInternalAdminUser())
	}

	gr := &multierror.Group{}
	for idx, consumerGroup := range d.consumers {
		cGroup := consumerGroup
		cGroupID := fmt.Sprintf("c-%d", idx)
		logger := d.logger.WithField("consumer", cGroupID)
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
			d.cancel()
			return err
		})
	}

	return gr.Wait().ErrorOrNil()
}

func (d *txSenderDaemon) Close() error {
	var gerr error
	for _, consumerGroup := range d.consumers {
		gerr = errors.CombineErrors(gerr, consumerGroup.Close())
	}

	return gerr
}

func readinessOpt(apiClient api.MetricClient, redisCli redis.Client, consumer messenger.Consumer) app.Option {
	return func(ap *app.App) error {
		ap.AddReadinessCheck("kafka", consumer.Checker)
		ap.AddReadinessCheck("api", apiClient.Checker())
		if redisCli != nil {
			ap.AddReadinessCheck("redis", redisCli.Ping)
		}
		return nil
	}
}
