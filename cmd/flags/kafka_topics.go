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
