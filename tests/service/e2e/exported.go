package e2e

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/consensys/orchestrate/src/api"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/utils"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	"github.com/consensys/orchestrate/src/infra/kafka/testutils"
	"github.com/consensys/orchestrate/tests/service/e2e/cucumber"
	"github.com/consensys/orchestrate/tests/service/e2e/cucumber/alias"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Start starts application
func Start(ctx context.Context, cfg *Config) error {
	logger := log.FromContext(ctx)
	logger.Info("starting execution...")

	var gerr error
	// Create context for application
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cctx = multitenancy.WithUserInfo(cctx, multitenancy.DefaultUser())

	rawTestData, _ := json.Marshal(cfg.TestData)
	initComponents(cctx, string(rawTestData))

	err := importTestIdentities(cctx, cfg.TestData)
	if err != nil {
		return err
	}

	chainUUIDs, err := initTestChains(cctx, cfg.TestData)
	if err != nil {
		return err
	}

	consumerTracker, err := testutils.NewConsumerTracker(cfg.KafkaCfg, &api.TopicConfig{})
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := consumerTracker.Consume(cctx, []string{cfg.TestData.NotificationsTopic})
		if err != nil {
			gerr = err
			log.FromContext(cctx).WithError(err).Error("error on consumer")
		}

		cancel()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		logger.WithField("tags", cfg.Cucumber.Tags).
			WithField("concurrency", cfg.Cucumber.Concurrency).
			WithField("paths", cfg.Cucumber.Paths).
			WithField("output", cfg.Cucumber.Output).
			Info("service ready")

		err := cucumber.Run(cfg.Cucumber, consumerTracker, cfg.TestData.NotificationsTopic, cfg.Timeout)
		if err != nil {
			gerr = err
			logger.WithError(err).Error("error on cucumber")
		}

		cancel()
		wg.Done()
	}()

	wg.Wait()

	time.Sleep(time.Second * 3) // Wait few seconds to allow ongoing work to complete
	if err := removeTestChains(ctx, chainUUIDs); err != nil {
		gerr = errors.CombineErrors(gerr, err)
	}

	return gerr
}

func Stop(ctx context.Context) error {
	log.FromContext(ctx).Info("Cucumber: stopping execution...")
	return nil
}

func initComponents(ctx context.Context, rawTestData string) {
	utils.InParallel(
		// Initialize logger
		func() {
			cfg := log.NewConfig(viper.GetViper())
			// Create and configure logger
			logger := logrus.StandardLogger()
			_ = log.ConfigureLogger(cfg, logger)
		},
		// Initialize cucumber handlers
		func() {
			alias.Init(rawTestData)
			client.Init()
			ethclient.Init(ctx)
		},
	)
}
