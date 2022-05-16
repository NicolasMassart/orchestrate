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
	txlistener "github.com/consensys/orchestrate/src/tx-listener"
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

func TxListenerFlags(f *pflag.FlagSet) {
	log.Flags(f)
	authkey.Flags(f)
	app.MetricFlags(f)
	orchestrateclient.Flags(f)

	KafkaFlags(f)
	KafkaConsumerFlags(f)
	KafkaTopicTxSender(f)
	KafkaTopicTxListener(f)

	metricregistry.Flags(f, tcpmetrics.ModuleName)
	providerRefreshInterval(f)
}

func providerRefreshInterval(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Time interval for refreshing the internal state such as active chains, pending jobs and subscriptions
Environment variable: %q`, providerRefreshIntervalEnv)
	f.Duration(providerRefreshIntervalFlag, providerRefreshIntervalDefault, desc)
	_ = viper.BindPFlag(providerRefreshIntervalViperKey, f.Lookup(providerRefreshIntervalFlag))
}

func NewTxListenerConfig(vipr *viper.Viper) *txlistener.Config {
	orchestrateAPIBackOff := backoff.IncrementalBackOffWithMaxRetries(time.Millisecond*500, time.Second, 5)

	httpClientCfg := http.NewDefaultConfig()
	httpClientCfg.XAPIKey = vipr.GetString(authkey.APIKeyViperKey)

	kafkaCfg := NewKafkaConfig(vipr)
	kafkaCfg.DisableCommitOnRead = true

	return &txlistener.Config{
		IsMultiTenancyEnabled: viper.GetBool(multitenancy.EnabledViperKey),
		App:                   app.NewConfig(vipr),
		ConsumerTopic:         viper.GetString(TxListenerViperKey),
		Messenger:             NewConsumerConfig(vipr),
		HTTPClient:            httpClientCfg,
		API:                   orchestrateclient.NewConfigFromViper(vipr, orchestrateAPIBackOff),
		RetryInterval:         vipr.GetDuration(providerRefreshIntervalViperKey),
		Kafka:                 kafkaCfg,
	}
}
