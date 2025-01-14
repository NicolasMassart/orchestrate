package txlistener

import (
	"context"
	"time"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	listenermetrics "github.com/consensys/orchestrate/src/tx-listener/tx-listener/metrics"
	"github.com/spf13/viper"
)

func New(ctx context.Context, cfg *Config) (*app.App, error) {
	// @TODO Refactor as part of https://github.com/ConsenSys/orchestrate/issues/637
	viper.Set(utils.RetryMaxIntervalViperKey, 30*time.Second)
	viper.Set(utils.RetryMaxElapsedTimeViperKey, 1*time.Hour)
	rpc.Init(ctx)

	apiClient := orchestrateclient.NewHTTPClient(http.NewClient(cfg.HTTPClient), cfg.API)

	kafkaProdClient, err := kafka.NewProducer(cfg.Kafka)
	if err != nil {
		return nil, err
	}

	var listenerMetrics listenermetrics.ListenerMetrics
	if cfg.App.Metrics.IsActive(listenermetrics.ModuleName) {
		listenerMetrics = listenermetrics.NewListenerMetrics()
	} else {
		listenerMetrics = listenermetrics.NewListenerNopMetrics()
	}

	return NewTxListener(
		cfg,
		kafkaProdClient,
		apiClient,
		rpc.GlobalClient(),
		listenerMetrics,
	)
}
