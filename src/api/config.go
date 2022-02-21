package api

import (
	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authjwt "github.com/consensys/orchestrate/pkg/toolkit/app/auth/jwt/jose"
	authkey "github.com/consensys/orchestrate/pkg/toolkit/app/auth/key"
	httpmetrics "github.com/consensys/orchestrate/pkg/toolkit/app/http/metrics"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	metricregistry "github.com/consensys/orchestrate/pkg/toolkit/app/metrics/registry"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	tcpmetrics "github.com/consensys/orchestrate/pkg/toolkit/tcp/metrics"
	"github.com/consensys/orchestrate/src/api/metrics"
	"github.com/consensys/orchestrate/src/api/proxy"
	store "github.com/consensys/orchestrate/src/api/store/multi"
	broker "github.com/consensys/orchestrate/src/infra/broker/sarama"
	quorumkeymanager "github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Flags register flags for API
func Flags(f *pflag.FlagSet) {
	flags.QKMFlags(f)

	log.Flags(f)
	multitenancy.Flags(f)
	authjwt.Flags(f)
	authkey.Flags(f)
	broker.KafkaProducerFlags(f)
	broker.KafkaTopicTxSender(f)
	store.Flags(f)
	app.Flags(f)
	app.MetricFlags(f)
	metricregistry.Flags(f, httpmetrics.ModuleName, tcpmetrics.ModuleName, metrics.ModuleName)
	proxy.Flags(f)
}

type Config struct {
	App          *app.Config
	Store        *store.Config
	Multitenancy bool
	Proxy        *proxy.Config
	QKM          *quorumkeymanager.Config
}

func NewConfig(vipr *viper.Viper) *Config {
	return &Config{
		App:          app.NewConfig(vipr),
		Store:        store.NewConfig(vipr),
		Multitenancy: viper.GetBool(multitenancy.EnabledViperKey),
		Proxy:        proxy.NewConfig(),
		QKM:          flags.NewQKMConfig(vipr),
	}
}
