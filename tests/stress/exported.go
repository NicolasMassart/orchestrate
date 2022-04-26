package stress

import (
	"context"
	"sync"
	"time"

	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/pkg/backoff"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	testutils2 "github.com/consensys/orchestrate/src/infra/notifier/kafka/testutils"
	"github.com/sirupsen/logrus"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/spf13/viper"
)

const component = "stress-test"

var (
	workload *WorkLoadService
	cancel   func()
)

// Start starts application
func Start(ctx context.Context) error {
	logger := log.WithContext(ctx).SetComponent(component)
	ctx = log.With(ctx, logger)
	logger.Info("starting execution...")

	var gerr error
	// Create context for application
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

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

	consumerTracker, err := testutils2.NewNotifierConsumerTracker(kafkaCfg, []string{cfg.KafkaTopic})
	if err != nil {
		return err
	}
	workload = NewService(cfg,
		consumerTracker,
		client,
		ethclient.GlobalClient())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := consumerTracker.Consume(ctx)
		if err != nil {
			gerr = errors.CombineErrors(gerr, err)
			logger.WithError(err).Error("error on consumer")
		}

		cancel()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		err := workload.Run(ctx)
		if err != nil {
			gerr = errors.CombineErrors(gerr, err)
			logger.WithError(err).Error("error on workload test")
		}

		cancel()
		wg.Done()
	}()

	wg.Wait()
	return gerr
}

func Stop(ctx context.Context) error {
	log.WithContext(ctx).Info("stopping stress test execution...")
	cancel()
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
