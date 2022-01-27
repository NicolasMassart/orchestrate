package utils

import (
	"github.com/consensys/orchestrate/src/infra/broker/sarama"
	"github.com/consensys/orchestrate/tests/utils"
)

var TOPICS = map[string]string{
	utils.TxSenderTopicKey:  sarama.TxSenderViperKey,
	utils.TxDecodedTopicKey: sarama.TxDecodedViperKey,
	utils.TxRecoverTopicKey: sarama.TxRecoverViperKey,
}
