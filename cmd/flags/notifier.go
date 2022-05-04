package flags

import (
	"fmt"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/src/notifier"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault(notifierMaxRetriesViperKey, notifierMaxRetriesDefault)
	_ = viper.BindEnv(notifierMaxRetriesViperKey, notifierMaxRetriesEnv)
}

const (
	notifierMaxRetriesFlag     = "notifier-max-retries"
	notifierMaxRetriesViperKey = "notifier.max.retries"
	notifierMaxRetriesDefault  = uint8(3)
	notifierMaxRetriesEnv      = "NOTIFIER_MAX_RETRIES"
)

func NotifierFlags(f *pflag.FlagSet) {
	KafkaTopicNotifier(f)
	notifierMaxRetries(f)
	orchestrateclient.Flags(f)
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
		MaxRetries:    int(viper.GetUint64(notifierMaxRetriesViperKey)),
	}
}
