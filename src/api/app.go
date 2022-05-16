package api

import (
	"reflect"
	"time"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/src/infra/messenger"

	service "github.com/consensys/orchestrate/src/api/service/listener"
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
)

func NewAPI(
	cfg *Config,
	db postgres.Client,
	jwt, key auth.Checker,
	keyManagerClient qkmclient.KeyManagerClient,
	qkmStoreID string,
	ec ethclient.Client,
	messengerClient sdk.OrchestrateMessenger,
	notifierDaemon app.Daemon,
) (*app.App, error) {
	// Metrics
	var appMetrics metrics.TransactionSchedulerMetrics
	if cfg.App.Metrics.IsActive(metrics.ModuleName) {
		appMetrics = metrics.NewTransactionSchedulerMetrics()
	} else {
		appMetrics = metrics.NewTransactionSchedulerNopMetrics()
	}

	ucs := builder.NewUseCases(
		postgresstore.New(db),
		appMetrics,
		keyManagerClient,
		qkmStoreID,
		ec,
		messengerClient,
	)

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

	subscriptionRouter := service.NewSubscriptionRouter(ucs.EventStreams().NotifyContractEvents())
	jobRouter := service.NewJobHandler(ucs.Jobs().Update())
	notificationRouter := service.NewNotificationHandler(ucs.Notifications().Ack())
	eventStreamRouter := service.NewEventStreamHandler(ucs.EventStreams().Update())

	msgConsumer, err := service.NewMessageConsumer(
		cfg.Kafka, []string{cfg.Messenger.TopicAPI},
		jobRouter, subscriptionRouter, notificationRouter, eventStreamRouter)
	if err != nil {
		return nil, err
	}

	// Create app
	appli, err := app.New(
		cfg.App,
		app.MultiTenancyOpt("auth", jwt, key, cfg.Multitenancy),
		ReadinessOpt(db, msgConsumer),
		app.MetricsOpt(appMetrics),
		accessLogMid,
		rateLimitOpt,
		apiHandlerOpt,
		httpCacheOpt,
		reverseProxyOpt,
		app.ProviderOpt(NewProvider(ucs.Chains().Search(), time.Second, cfg.Proxy.ProxyCacheTTL)),
	)
	if err != nil {
		return nil, err
	}

	appli.RegisterDaemon(NewConsumerService(msgConsumer))
	appli.RegisterDaemon(notifierDaemon)

	return appli, nil
}

func ReadinessOpt(postgresClient postgres.Client, consumer messenger.Consumer) app.Option {
	return func(ap *app.App) error {
		ap.AddReadinessCheck("database", func() error { return postgresClient.Exec("SELECT 1") })
		ap.AddReadinessCheck("kafka", consumer.Checker)
		return nil
	}
}
