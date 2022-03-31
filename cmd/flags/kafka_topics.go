package flags

import (
	"fmt"

	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	// Kafka topics for the tx workflow
	viper.SetDefault(TxSenderViperKey, txSenderTopicDefault)
	_ = viper.BindEnv(TxSenderViperKey, txSenderTopicEnv)
	viper.SetDefault(TxDecodedViperKey, txDecodedTopicDefault)
	_ = viper.BindEnv(TxDecodedViperKey, txDecodedTopicEnv)
	viper.SetDefault(TxRecoverViperKey, txRecoverTopicDefault)
	_ = viper.BindEnv(TxRecoverViperKey, txRecoverTopicEnv)
}

const (
	txSenderFlag         = "topic-tx-sender"
	TxSenderViperKey     = "topic.tx.sender"
	txSenderTopicEnv     = "TOPIC_TX_SENDER"
	txSenderTopicDefault = kafka.DefaultTxSenderTopic

	txDecodedFlag         = "topic-tx-decoded"
	TxDecodedViperKey     = "topic.tx.decoded"
	txDecodedTopicEnv     = "TOPIC_TX_DECODED"
	txDecodedTopicDefault = kafka.DefaultTxDecodedTopic

	txRecoverFlag         = "topic-tx-recover"
	TxRecoverViperKey     = "topic.tx.recover"
	txRecoverTopicEnv     = "TOPIC_TX_RECOVER"
	txRecoverTopicDefault = kafka.DefaultTxRecoverTopic
)

// KafkaTopicTxSender register flag for Kafka topic
func KafkaTopicTxSender(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Topic for messages between the API and the Tx-Sender.
Environment variable: %q`, txSenderTopicEnv)
	f.String(txSenderFlag, txSenderTopicDefault, desc)
	_ = viper.BindPFlag(TxSenderViperKey, f.Lookup(txSenderFlag))
}

// KafkaTopicTxRecover register flag for Kafka topic
func KafkaTopicTxRecover(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Topic for failed transaction messages.
Environment variable: %q`, txRecoverTopicEnv)
	f.String(txRecoverFlag, txRecoverTopicDefault, desc)
	_ = viper.BindPFlag(TxRecoverViperKey, f.Lookup(txRecoverFlag))
}

// KafkaTopicTxDecoded register flag for Kafka topic
func KafkaTopicTxDecoded(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Topic for successful transaction messages (receipts).
Environment variable: %q`, txDecodedTopicEnv)
	f.String(txDecodedFlag, txDecodedTopicDefault, desc)
	_ = viper.BindPFlag(TxDecodedViperKey, f.Lookup(txDecodedFlag))
}

func NewKafkaTopicConfig(vipr *viper.Viper) *kafka.TopicConfig {
	return &kafka.TopicConfig{
		Sender:  vipr.GetString(TxSenderViperKey),
		Decoded: vipr.GetString(TxDecodedViperKey),
		Recover: vipr.GetString(TxRecoverViperKey),
	}
}
