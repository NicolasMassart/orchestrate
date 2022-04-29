package flags

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	// Kafka topics for the tx workflow
	viper.SetDefault(TxSenderViperKey, txSenderTopicDefault)
	_ = viper.BindEnv(TxSenderViperKey, txSenderTopicEnv)
	viper.SetDefault(TxListenerViperKey, txListenerTopicDefault)
	_ = viper.BindEnv(TxListenerViperKey, txListenerTopicEnv)
}

const (
	txSenderFlag         = "topic-tx-sender"
	TxSenderViperKey     = "topic.tx.sender"
	txSenderTopicEnv     = "TOPIC_TX_SENDER"
	txSenderTopicDefault = "topic-tx-sender"
)

// KafkaTopicTxSender register flag for Kafka topic
func KafkaTopicTxSender(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Topic for messages between the API and the Tx-Sender.
Environment variable: %q`, txSenderTopicEnv)
	f.String(txSenderFlag, txSenderTopicDefault, desc)
	_ = viper.BindPFlag(TxSenderViperKey, f.Lookup(txSenderFlag))
}

const (
	txListenerFlag         = "topic-tx-listener"
	TxListenerViperKey     = "topic.tx.listener"
	txListenerTopicEnv     = "TOPIC_TX_LISTENER"
	txListenerTopicDefault = "topic-tx-listener"
)

func KafkaTopicTxListener(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Topic for messages between the API and the Tx-Listener.
Environment variable: %q`, txListenerTopicEnv)
	f.String(txListenerFlag, txListenerTopicDefault, desc)
	_ = viper.BindPFlag(TxListenerViperKey, f.Lookup(txListenerFlag))
}

const (
	notifierTopicFlag     = "topic-notifier"
	NotifierTopicViperKey = "topic.notifier"
	notifierTopicEnv      = "TOPIC_NOTIFIER"
	notifierTopicDefault  = "topic-notifier"
)

func KafkaTopicNotifier(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Topic for messages between the API and the Notifier MS.
Environment variable: %q`, notifierTopicEnv)
	f.String(notifierTopicFlag, notifierTopicDefault, desc)
	_ = viper.BindPFlag(NotifierTopicViperKey, f.Lookup(notifierTopicFlag))
}
