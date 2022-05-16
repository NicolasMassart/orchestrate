package kafka

import (
	"bytes"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/src/entities"
	infra "github.com/consensys/orchestrate/src/infra/api"
)

func NewMessage(msg *sarama.ConsumerMessage, session sarama.ConsumerGroupSession) (*entities.Message, error) {
	reqMsg := &entities.Message{}

	err := infra.UnmarshalBody(bytes.NewReader(msg.Value), reqMsg)
	if err != nil {
		return nil, err
	}

	reqMsg.Offset = msg.Offset
	reqMsg.Commit = func() error {
		session.MarkMessage(msg, "")
		session.Commit()
		return nil
	}

	return reqMsg, nil
}
