package kafka

import (
	broker "github.com/consensys/orchestrate/src/infra/broker/sarama"
	"github.com/spf13/viper"
)

type Config struct {
	OutTopic string
}

func NewConfig() *Config {
	return &Config{
		OutTopic: viper.GetString(broker.TxDecodedViperKey),
	}
}
