package messenger

import (
	"encoding/json"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/kafka"
)

type ProducerClient struct {
	client kafka.Producer
	cfg    *Config
}

var _ sdk.OrchestrateMessenger = &ProducerClient{}

func NewProducerClient(cfg *Config, client kafka.Producer) *ProducerClient {
	return &ProducerClient{
		client: client,
		cfg:    cfg,
	}
}

func (c *ProducerClient) sendMessage(topic string, msgType entities.RequestMessageType, msgBody interface{}, partitionKey string, userInfo *multitenancy.UserInfo) error {
	if topic == "" {
		return errors.InvalidParameterError("topic not defined")
	}

	bBody, err := json.Marshal(msgBody)
	if err != nil {
		return errors.EncodingError("failed to marshall consumer message body")
	}

	headers := map[string]interface{}{}
	if userInfo.AuthMode == multitenancy.AuthMethodJWT || userInfo.AuthMode == multitenancy.AuthMethodAPIKey {
		headers[utils.UserInfoHeader] = userInfo
	}

	err = c.client.Send(&entities.Message{
		Type: msgType,
		Body: bBody,
	}, topic, partitionKey, headers)

	if err != nil {
		return err
	}

	return nil
}
