package api

import (
	"reflect"
	"time"

	"github.com/consensys/orchestrate/src/infra/kafka"

	"github.com/consensys/orchestrate/src/infra/postgres"

	"github.com/consensys/orchestrate/pkg/toolkit/app/http/middleware/httpcache"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http/middleware/ratelimit"
	"github.com/consensys/orchestrate/src/api/proxy"
	"github.com/dgraph-io/ristretto"

	postgresstore "github.com/consensys/orchestrate/src/api/store/postgres"
	"github.com/consensys/orchestrate/src/infra/ethclient"

	qkmclient "github.com/consensys/quorum-key-manager/pkg/client"

	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/pkg/toolkit/app/auth"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http/config/dynamic"
	pkgproxy "github.com/consensys/orchestrate/pkg/toolkit/app/http/handler/proxy"
	"github.com/consensys/orchestrate/src/api/business/builder"
	"github.com/consensys/orchestrate/src/api/metrics"
	"github.com/consensys/orchestrate/src/api/service/controllers"
	broker "github.com/consensys/orchestrate/src/infra/kafka/sarama"
)

func NewAPI(
	cfg *Config,
	db postgres.Client,
	jwt, key auth.Checker,
	keyManagerClient qkmclient.KeyManagerClient,
	qkmStoreID string,
	ec ethclient.Client,
	syncProducer kafka.Producer,
	topicCfg *broker.TopicConfig,
) (*app.App, error) {
	// Metrics
	var appMetrics metrics.TransactionSchedulerMetrics
	if cfg.App.Metrics.IsActive(metrics.ModuleName) {
		appMetrics = metrics.NewTransactionSchedulerMetrics()
	} else {
		appMetrics = metrics.NewTransactionSchedulerNopMetrics()
	}

	ucs := builder.NewUseCases(postgresstore.New(db), appMetrics, keyManagerClient, qkmStoreID, ec, syncProducer, topicCfg)

	// Option of the API
	apiHandlerOpt := app.HandlerOpt(reflect.TypeOf(&dynamic.API{}), controllers.NewBuilder(ucs, keyManagerClient, qkmStoreID))

	// ReverseProxy Handler
	proxyBuilder, err := pkgproxy.NewBuilder(cfg.Proxy.ServersTransport, nil)
	if err != nil {
		return nil, err
	}
	reverseProxyOpt := app.HandlerOpt(
		reflect.TypeOf(&dynamic.ReverseProxy{}),
		proxyBuilder,
	)

	cache, err := ristretto.NewCache(cfg.Proxy.Cache)
	if err != nil {
		return nil, err
	}

	// RateLimit Middleware
	rateLimitOpt := app.MiddlewareOpt(
		reflect.TypeOf(&dynamic.RateLimit{}),
		ratelimit.NewBuilder(ratelimit.NewManager(cache)),
	)

	// HTTPCache Middleware
	httpCacheOpt := app.MiddlewareOpt(
		reflect.TypeOf(&dynamic.HTTPCache{}),
		httpcache.NewBuilder(cache, proxy.HTTPCacheRequest, proxy.HTTPCacheResponse),
	)

	var accessLogMid app.Option
	if cfg.App.HTTP.AccessLog {
		accessLogMid = app.LoggerMiddlewareOpt("base")
	} else {
		accessLogMid = app.NonOpt()
	}

	// Create app
	return app.New(
		cfg.App,
		app.MultiTenancyOpt("auth", jwt, key, cfg.Multitenancy),
		ReadinessOpt(db, syncProducer.Client()),
		app.MetricsOpt(appMetrics),
		accessLogMid,
		rateLimitOpt,
		apiHandlerOpt,
		httpCacheOpt,
		reverseProxyOpt,
		app.ProviderOpt(NewProvider(ucs.SearchChains(), time.Second, cfg.Proxy.ProxyCacheTTL)),
	)
}

func ReadinessOpt(postgresClient postgres.Client, brokerClient kafka.Client) app.Option {
	return func(ap *app.App) error {
		ap.AddReadinessCheck("database", func() error { return postgresClient.Exec("SELECT 1") })
		ap.AddReadinessCheck("kafka", brokerClient.Checker)
		return nil
	}
}
