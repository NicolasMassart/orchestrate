package flags

import (
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authjwt "github.com/consensys/orchestrate/pkg/toolkit/app/auth/jwt/jose"
	authkey "github.com/consensys/orchestrate/pkg/toolkit/app/auth/key"
	httpmetrics "github.com/consensys/orchestrate/pkg/toolkit/app/http/metrics"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	metricregistry "github.com/consensys/orchestrate/pkg/toolkit/app/metrics/registry"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	tcpmetrics "github.com/consensys/orchestrate/pkg/toolkit/tcp/metrics"
	"github.com/consensys/orchestrate/src/api"
	"github.com/consensys/orchestrate/src/api/metrics"
	"github.com/consensys/orchestrate/src/api/proxy"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func NewAPIFlags(f *pflag.FlagSet) {
	QKMFlags(f)
	PGFlags(f)

	KafkaFlags(f)
	KafkaConsumerFlags(f)
	KafkaTopicTxSender(f)
	KafkaTopicTxListener(f)
	KafkaTopicAPI(f)

	log.Flags(f)
	multitenancy.Flags(f)
	authjwt.Flags(f)
	authkey.Flags(f)
	app.Flags(f)
	app.MetricFlags(f)
	metricregistry.Flags(f, httpmetrics.ModuleName, tcpmetrics.ModuleName, metrics.ModuleName)
	proxy.Flags(f)
}

func NewAPIConfig(vipr *viper.Viper) *api.Config {
	return &api.Config{
		App:      app.NewConfig(vipr),
		Postgres: NewPGConfig(vipr),
		Kafka:    NewKafkaConfig(vipr),
		KafkaTopics: &api.TopicConfig{
			API:      viper.GetString(APITopicViperKey),
			Sender:   viper.GetString(TxSenderViperKey),
			Listener: viper.GetString(TxListenerViperKey),
			Notifier: viper.GetString(NotifierTopicViperKey),
		},
		Multitenancy: vipr.GetBool(multitenancy.EnabledViperKey),
		Proxy:        proxy.NewConfig(),
		QKM:          NewQKMConfig(vipr),
	}
}
