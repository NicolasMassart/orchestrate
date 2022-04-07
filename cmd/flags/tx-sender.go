package flags

import (
	"fmt"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	txsender "github.com/consensys/orchestrate/src/tx-sender"

	"github.com/cenkalti/backoff/v4"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authkey "github.com/consensys/orchestrate/pkg/toolkit/app/auth/key"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	metricregistry "github.com/consensys/orchestrate/pkg/toolkit/app/metrics/registry"
	tcpmetrics "github.com/consensys/orchestrate/pkg/toolkit/tcp/metrics"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault(TxSenderMetricsURLViperKey, txSenderMetricsURLDefault)
	_ = viper.BindEnv(TxSenderMetricsURLViperKey, txSenderMetricsURLEnv)

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
	TxSenderMetricsURLViperKey = "tx-sender.metrics.url"
	txSenderMetricsURLDefault  = "localhost:8082"
	txSenderMetricsURLEnv      = "TX_SENDER_METRICS_URL"
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

func TxSenderFlags(f *pflag.FlagSet) {
	RedisFlags(f)
	QKMFlags(f)

	KafkaFlags(f)
	KafkaConsumerFlags(f)
	KafkaTopicTxSender(f)

	log.Flags(f)
	authkey.Flags(f)
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
Environment variable: %q`, []string{txsender.NonceManagerTypeInMemory, txsender.NonceManagerTypeRedis}, nonceManagerTypeEnv)
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

func retryMessageBackOff() backoff.BackOff {
	bckOff := backoff.NewExponentialBackOff()
	bckOff.MaxInterval = time.Second * 5
	bckOff.MaxElapsedTime = time.Second * 30
	return bckOff
}

func NewTxSenderConfig(vipr *viper.Viper) *txsender.Config {
	return &txsender.Config{
		App:                    app.NewConfig(vipr),
		Kafka:                  NewKafkaConfig(vipr),
		ConsumerTopic:          viper.GetString(TxSenderViperKey),
		ProxyURL:               vipr.GetString(orchestrateclient.URLViperKey),
		NonceMaxRecovery:       vipr.GetUint64(NonceMaxRecoveryViperKey),
		BckOff:                 retryMessageBackOff(),
		NonceManagerType:       vipr.GetString(nonceManagerTypeViperKey),
		NonceManagerExpiration: vipr.GetDuration(NonceManagerExpirationViperKey),
		RedisCfg:               NewRedisConfig(vipr),
		NConsumer:              int(vipr.GetUint64(KafkaConsumerViperKey)),
		IsMultiTenancyEnabled:  vipr.GetBool(multitenancy.EnabledViperKey),
		QKM:                    NewQKMConfig(vipr),
	}
}
