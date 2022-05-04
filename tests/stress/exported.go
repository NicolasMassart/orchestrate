package stress

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/tests/pkg/trackers"

	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/pkg/backoff"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	"github.com/sirupsen/logrus"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/spf13/viper"
)

const component = "stress-test"

var (
	workload *WorkLoadService
)

// Start starts application
func Start(ctx context.Context) error {
	logger := log.WithContext(ctx).SetComponent(component)
	ctx = log.With(ctx, logger)
	logger.Info("starting execution...")

	var gerr error

	initComponents(ctx)

	cfg, err := InitConfig(viper.GetViper())
	if err != nil {
		return err
	}

	kafkaCfg := flags.NewKafkaConfig(viper.GetViper())
	httpClientCfg := http.NewConfig(viper.GetViper())
	httpClient := http.NewClient(httpClientCfg)
	backoffConf := orchestrateclient.NewConfigFromViper(viper.GetViper(),
		backoff.IncrementalBackOff(time.Second, time.Second*5, time.Minute))
	client := orchestrateclient.NewHTTPClient(httpClient, backoffConf)

	consumerTracker, err := trackers.NewNotifierConsumerTacker(*kafkaCfg, []string{cfg.KafkaTopic})
	if err != nil {
		return err
	}
	workload = NewService(cfg,
		consumerTracker,
		client,
		ethclient.GlobalClient())

	go func() {
		err := consumerTracker.Consume(ctx)
		if err != nil {
			gerr = errors.CombineErrors(gerr, err)
			logger.WithError(err).Error("error on consumer")
		}
	}()

	err = workload.Run(ctx)
	if err != nil {
		gerr = errors.CombineErrors(gerr, err)
		logger.WithError(err).Error("error on workload test")
	}

	_ = consumerTracker.Close()
	return gerr
}

func Stop(ctx context.Context) error {
	log.WithContext(ctx).Info("stopping stress test execution...")
	return nil
}

func initComponents(ctx context.Context) {
	utils.InParallel(
		func() {
			ethclient.Init(ctx)
		},
		// Initialize logger
		func() {
			cfg := log.NewConfig(viper.GetViper())
			// Create and configure logger
			logger := logrus.StandardLogger()
			_ = log.ConfigureLogger(cfg, logger)
		},
	)
}
