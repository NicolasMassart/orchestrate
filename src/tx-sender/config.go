package txsender

import (
	"fmt"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	quorumkeymanager "github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"

	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/src/infra/redis/redigo"

	authkey "github.com/consensys/orchestrate/pkg/toolkit/app/auth/key"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	metricregistry "github.com/consensys/orchestrate/pkg/toolkit/app/metrics/registry"
	tcpmetrics "github.com/consensys/orchestrate/pkg/toolkit/tcp/metrics"
	broker "github.com/consensys/orchestrate/src/infra/broker/sarama"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault(MetricsURLViperKey, metricsURLDefault)
	_ = viper.BindEnv(MetricsURLViperKey, metricsURLEnv)

	viper.SetDefault(NonceMaxRecoveryViperKey, nonceMaxRecoveryDefault)
	_ = viper.BindEnv(NonceMaxRecoveryViperKey, nonceMaxRecoveryEnv)

	viper.SetDefault(nonceManagerTypeViperKey, nonceManagerTypeDefault)
	_ = viper.BindEnv(nonceManagerTypeViperKey, nonceManagerTypeEnv)

	viper.SetDefault(NonceManagerExpirationViperKey, nonceManagerExpirationDefault)
	_ = viper.BindEnv(NonceManagerExpirationViperKey, nonceManagerExpirationEnv)

	viper.SetDefault(KafkaConsumerViperKey, kafkaConsumerDefault)
	_ = viper.BindEnv(KafkaConsumerViperKey, KafkaConsumerEnv)
}

const (
	MetricsURLViperKey = "tx-sender.metrics.url"
	metricsURLDefault  = "localhost:8082"
	metricsURLEnv      = "TX_SENDER_METRICS_URL"
)

const (
	nonceMaxRecoveryFlag     = "nonce-max-recovery"
	NonceMaxRecoveryViperKey = "nonce.max.recovery"
	nonceMaxRecoveryDefault  = 5
	nonceMaxRecoveryEnv      = "NONCE_MAX_RECOVERY"
)

const (
	nonceManagerTypeFlag     = "nonce-manager-type"
	nonceManagerTypeViperKey = "nonce.manager.type"
	nonceManagerTypeDefault  = "redis"
	nonceManagerTypeEnv      = "NONCE_MANAGER_TYPE"

	NonceManagerTypeInMemory = "in-memory"
	NonceManagerTypeRedis    = "redis"
)

const (
	nonceManagerExpirationFlag     = "nonce-manager-expiration"
	NonceManagerExpirationViperKey = "nonce.manager.expiration"
	nonceManagerExpirationDefault  = 5 * time.Minute
	nonceManagerExpirationEnv      = "NONCE_MANAGER_EXPIRATION"
)

const (
	kafkaConsumersFlag    = "kafka-consumers"
	KafkaConsumerViperKey = "kafka.consumers"
	kafkaConsumerDefault  = uint8(1)
	KafkaConsumerEnv      = "KAFKA_NUM_CONSUMERS"
)

// Flags register flags for tx sentry
func Flags(f *pflag.FlagSet) {
	flags.RedisFlags(f)
	flags.QKMFlags(f)

	log.Flags(f)
	authkey.Flags(f)
	broker.KafkaConsumerFlags(f)
	broker.KafkaTopicTxSender(f)
	broker.KafkaTopicTxRecover(f)
	orchestrateclient.Flags(f)
	app.MetricFlags(f)
	metricregistry.Flags(f, tcpmetrics.ModuleName)

	maxRecovery(f)
	nonceManagerType(f)
	nonceManagerExpiration(f)
	kafkaConsumers(f)
}

func maxRecovery(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Maximum number of times to try to recover a transaction with invalid nonce before returning an error.
Environment variable: %q`, nonceMaxRecoveryEnv)
	f.Int(nonceMaxRecoveryFlag, nonceMaxRecoveryDefault, desc)
	_ = viper.BindPFlag(NonceMaxRecoveryViperKey, f.Lookup(nonceMaxRecoveryFlag))
}

func nonceManagerType(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Type of Nonce manager cache (one of %q)
Environment variable: %q`, []string{NonceManagerTypeInMemory, NonceManagerTypeRedis}, nonceManagerTypeEnv)
	f.String(nonceManagerTypeFlag, nonceManagerTypeDefault, desc)
	_ = viper.BindPFlag(nonceManagerTypeViperKey, f.Lookup(nonceManagerTypeFlag))
}

func nonceManagerExpiration(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Nonce manager cache expiration time (TTL).
Environment variable: %q`, nonceManagerExpirationEnv)
	f.Duration(nonceManagerExpirationFlag, nonceManagerExpirationDefault, desc)
	_ = viper.BindPFlag(NonceManagerExpirationViperKey, f.Lookup(nonceManagerExpirationFlag))
}

func kafkaConsumers(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Number of parallel kafka consumers to initialize.
Environment variable: %q`, KafkaConsumerEnv)
	f.Uint8(kafkaConsumersFlag, kafkaConsumerDefault, desc)
	_ = viper.BindPFlag(KafkaConsumerViperKey, f.Lookup(kafkaConsumersFlag))
}

type Config struct {
	App                    *app.Config
	GroupName              string
	NConsumer              int
	RecoverTopic           string
	SenderTopic            string
	ProxyURL               string
	BckOff                 backoff.BackOff
	NonceMaxRecovery       uint64
	NonceManagerType       string
	IsMultiTenancyEnabled  bool
	RedisCfg               *redigo.Config
	NonceManagerExpiration time.Duration
	QKM                    *quorumkeymanager.Config
}

func NewConfig(vipr *viper.Viper) *Config {
	return &Config{
		App:                    app.NewConfig(vipr),
		GroupName:              vipr.GetString(broker.ConsumerGroupNameViperKey),
		RecoverTopic:           vipr.GetString(broker.TxRecoverViperKey),
		SenderTopic:            vipr.GetString(broker.TxSenderViperKey),
		ProxyURL:               vipr.GetString(orchestrateclient.URLViperKey),
		NonceMaxRecovery:       vipr.GetUint64(NonceMaxRecoveryViperKey),
		BckOff:                 retryMessageBackOff(),
		NonceManagerType:       vipr.GetString(nonceManagerTypeViperKey),
		NonceManagerExpiration: vipr.GetDuration(NonceManagerExpirationViperKey),
		RedisCfg:               flags.NewRedisConfig(vipr),
		NConsumer:              int(vipr.GetUint64(KafkaConsumerViperKey)),
		IsMultiTenancyEnabled:  vipr.GetBool(multitenancy.EnabledViperKey),
		QKM:                    flags.NewQKMConfig(vipr),
	}
}

func retryMessageBackOff() backoff.BackOff {
	bckOff := backoff.NewExponentialBackOff()
	bckOff.MaxInterval = time.Second * 15
	bckOff.MaxElapsedTime = time.Minute
	return bckOff
}
