package flags

import (
	"fmt"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/src/notifier"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault(KafkaConsumerViperKey, kafkaConsumerDefault)
	_ = viper.BindEnv(KafkaConsumerViperKey, KafkaConsumerEnv)

	viper.SetDefault(KafkaConsumerViperKey, kafkaConsumerDefault)
	_ = viper.BindEnv(KafkaConsumerViperKey, KafkaConsumerEnv)
}

const (
	notifierKafkaConsumersFlag    = "notifier-kafka-consumers"
	notifierKafkaConsumerViperKey = "notifier.kafka.consumers"
	notifierKafkaConsumerDefault  = uint8(1)
	notifierKafkaConsumerEnv      = "NOTIFIER_KAFKA_NUM_CONSUMERS"
)

const (
	notifierMaxRetriesFlag     = "notifier-max-retries"
	notifierMaxRetriesViperKey = "notifier.max.retries"
	notifierMaxRetriesDefault  = uint8(3)
	notifierMaxRetriesEnv      = "NOTIFIER_MAX_RETRIES"
)

func NotifierFlags(f *pflag.FlagSet) {
	KafkaTopicNotifier(f)
	notifierKafkaConsumers(f)
	notifierMaxRetries(f)
	orchestrateclient.Flags(f)
}

func notifierKafkaConsumers(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Number of parallel kafka consumers to initialize.
Environment variable: %q`, notifierKafkaConsumerEnv)
	f.Uint8(notifierKafkaConsumersFlag, notifierKafkaConsumerDefault, desc)
	_ = viper.BindPFlag(notifierKafkaConsumerViperKey, f.Lookup(notifierKafkaConsumersFlag))
}

func notifierMaxRetries(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Maximum number of retries before an event stream is suspended.
Environment variable: %q`, notifierMaxRetriesEnv)
	f.Uint8(notifierMaxRetriesFlag, notifierMaxRetriesDefault, desc)
	_ = viper.BindPFlag(notifierMaxRetriesViperKey, f.Lookup(notifierMaxRetriesFlag))
}

func NewNotifierConfig(vipr *viper.Viper) *notifier.Config {
	return &notifier.Config{
		Kafka:         NewKafkaConfig(vipr),
		ConsumerTopic: viper.GetString(NotifierTopicViperKey),
		NConsumer:     int(vipr.GetUint64(notifierKafkaConsumerViperKey)),
		MaxRetries:    int(viper.GetUint64(notifierMaxRetriesViperKey)),
	}
}
