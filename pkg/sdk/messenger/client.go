package messenger

import (
	"encoding/json"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/infra/kafka"
	"github.com/consensys/orchestrate/src/infra/messenger"
	types "github.com/consensys/orchestrate/src/infra/messenger/kafka"
)

type ProducerClient struct {
	client kafka.Producer
	topic  string
}

var _ sdk.OrchestrateMessenger = &ProducerClient{}

func NewProducerClient(client kafka.Producer, topic string) *ProducerClient {
	return &ProducerClient{
		client: client,
		topic:  topic,
	}
}

func (c *ProducerClient) sendMessage(msgType messenger.ConsumerRequestMessageType, msgBody interface{}, partitionKey string, userInfo *multitenancy.UserInfo) error {
	bBody, err := json.Marshal(msgBody)
	if err != nil {
		return errors.EncodingError("failed to marshall consumer message body")
	}

	headers := map[string]interface{}{}
	if userInfo.AuthMode == multitenancy.AuthMethodJWT {
		headers[utils.UserInfoHeader] = userInfo
	}

	err = c.client.Send(&types.ConsumerRequestMessage{
		Type: msgType,
		Body: bBody,
	}, c.topic, partitionKey, headers)

	if err != nil {
		return err
	}

	return nil
}
