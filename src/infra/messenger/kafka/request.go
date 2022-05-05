package kafka

import (
	"github.com/consensys/orchestrate/src/infra/messenger"
)

type ConsumerRequestMessage struct {
	Type messenger.ConsumerRequestMessageType
	Body []byte
}
