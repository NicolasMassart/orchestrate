package kafka

import (
	encoding "encoding/json"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
)

func DecodeJobMessage(rawMsg *sarama.ConsumerMessage) (interface{}, error) {
	job := &entities.Job{}
	err := encoding.Unmarshal(rawMsg.Value, job)
	if err != nil {
		errMessage := "failed to decode job message"
		return nil, errors.EncodingError(errMessage)
	}
	return job, nil
}
