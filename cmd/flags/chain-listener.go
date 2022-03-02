package flags

import (
	"fmt"
	"time"

	"github.com/consensys/orchestrate/pkg/backoff"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authkey "github.com/consensys/orchestrate/pkg/toolkit/app/auth/key"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	metricregistry "github.com/consensys/orchestrate/pkg/toolkit/app/metrics/registry"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	tcpmetrics "github.com/consensys/orchestrate/pkg/toolkit/tcp/metrics"
	chainlistener "github.com/consensys/orchestrate/src/chain-listener"
	broker "github.com/consensys/orchestrate/src/infra/broker/sarama"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault(providerRefreshIntervalViperKey, providerRefreshIntervalDefault)
	_ = viper.BindEnv(providerRefreshIntervalViperKey, providerRefreshIntervalEnv)
}

const (
	providerRefreshIntervalFlag     = "tx-listener-provider-refresh-interval"
	providerRefreshIntervalViperKey = "tx-listener-provider.refresh-interval"
	providerRefreshIntervalDefault  = time.Second
	providerRefreshIntervalEnv      = "TX_LISTENER_REFRESH_INTERVAL"
)

// sentryRefreshInterval register flags for API
func ChainListenerFlags(f *pflag.FlagSet) {
	log.Flags(f)
	authkey.Flags(f)
	txListenerFlags(f)
	metricregistry.Flags(f, tcpmetrics.ModuleName)
	providerRefreshInterval(f)
}

// sentryRefreshInterval register flags for API
func txListenerFlags(f *pflag.FlagSet) {
	broker.KafkaProducerFlags(f)
	broker.KafkaTopicTxDecoded(f)
	app.MetricFlags(f)
	orchestrateclient.Flags(f)
}

// ProviderRefreshInterval register flag for refresh interval duration
func providerRefreshInterval(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Time interval for refreshing the internal state such as active chains, pending jobs and subscriptions
Environment variable: %q`, providerRefreshIntervalEnv)
	f.Duration(providerRefreshIntervalFlag, providerRefreshIntervalDefault, desc)
	_ = viper.BindPFlag(providerRefreshIntervalViperKey, f.Lookup(providerRefreshIntervalFlag))
}

func NewChainListenerConfig(vipr *viper.Viper) *chainlistener.Config {
	orchestrateAPIBackOff := backoff.IncrementalBackOffWithMaxRetries(time.Millisecond*500, time.Second, 5)

	httpClientCfg := http.NewDefaultConfig()
	httpClientCfg.XAPIKey = vipr.GetString(authkey.APIKeyViperKey)

	chainListenerCfg := chainlistener.NewTxListenerConfig(
		vipr.GetDuration(providerRefreshIntervalViperKey),
		vipr.GetString(broker.TxDecodedViperKey))

	return &chainlistener.Config{
		IsMultiTenancyEnabled: viper.GetBool(multitenancy.EnabledViperKey),
		App:                   app.NewConfig(vipr),
		HTTPClient:            httpClientCfg,
		API:                   orchestrateclient.NewConfigFromViper(vipr, orchestrateAPIBackOff),
		ChainListenerConfig:   chainListenerCfg,
	}
}
