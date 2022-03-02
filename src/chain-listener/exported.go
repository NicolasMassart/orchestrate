package chainlistener

import (
	"context"
	"time"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/utils"
	listenermetrics "github.com/consensys/orchestrate/src/chain-listener/chain-listener/metrics"
	pkgsarama "github.com/consensys/orchestrate/src/infra/broker/sarama"
	"github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	"github.com/spf13/viper"
)

func New(ctx context.Context, cfg *Config) (*Service, error) {
	// @TODO Refactor as part of https://github.com/ConsenSys/orchestrate/issues/637
	viper.Set(utils.RetryMaxIntervalViperKey, 30*time.Second)
	viper.Set(utils.RetryMaxElapsedTimeViperKey, 1*time.Hour)
	rpc.Init(ctx)
	pkgsarama.InitSyncProducer(ctx)

	apiClient := orchestrateclient.NewHTTPClient(http.NewClient(cfg.HTTPClient), cfg.API)

	var listenerMetrics listenermetrics.ListenerMetrics
	if cfg.App.Metrics.IsActive(listenermetrics.ModuleName) {
		listenerMetrics = listenermetrics.NewListenerMetrics()
	} else {
		listenerMetrics = listenermetrics.NewListenerNopMetrics()
	}

	return NewChainListener(
		cfg,
		apiClient,
		pkgsarama.GlobalSyncProducer(),
		rpc.GlobalClient(),
		listenerMetrics,
	)
}
