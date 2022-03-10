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
	"github.com/consensys/orchestrate/src/api/metrics"
	"github.com/consensys/orchestrate/src/api/proxy"
	broker "github.com/consensys/orchestrate/src/infra/broker/sarama"
	"github.com/spf13/pflag"
)

func NewAPIFlags(f *pflag.FlagSet) {
	QKMFlags(f)
	PGFlags(f)

	log.Flags(f)
	multitenancy.Flags(f)
	authjwt.Flags(f)
	authkey.Flags(f)
	broker.KafkaProducerFlags(f)
	broker.KafkaTopicTxSender(f)
	app.Flags(f)
	app.MetricFlags(f)
	metricregistry.Flags(f, httpmetrics.ModuleName, tcpmetrics.ModuleName, metrics.ModuleName)
	proxy.Flags(f)
}
