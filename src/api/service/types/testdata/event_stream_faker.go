package testdata

import (
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/quorum-key-manager/pkg/common"
)

func FakeCreateWebhookEventStreamRequest() *api.CreateWebhookEventStreamRequest {
	return &api.CreateWebhookEventStreamRequest{
		Name:  "eventstream-webhook-" + common.RandString(5),
		Chain: "mainnet",
		URL:   "https://my-webhook-endpoint.com",
		Headers: map[string]string{
			"Authorization": "Bearer JWT",
		},
		Labels: map[string]string{
			"label": "labelValue",
		},
	}
}

func FakeCreateKafkaEventStreamRequest() *api.CreateKafkaEventStreamRequest {
	return &api.CreateKafkaEventStreamRequest{
		Name:  "eventstream-webhook-" + common.RandString(5),
		Chain: "mainnet",
		Topic: "topic-tx-decoded",
		Labels: map[string]string{
			"label": "labelValue",
		},
	}
}

func FakeUpdateWebhookEventStreamRequest() *api.UpdateWebhookEventStreamRequest {
	return &api.UpdateWebhookEventStreamRequest{
		Status: string(entities.EventStreamStatusPaused),
	}
}

func FakeUpdateKafkaEventStreamRequest() *api.UpdateKafkaEventStreamRequest {
	return &api.UpdateKafkaEventStreamRequest{
		Status: string(entities.EventStreamStatusPaused),
	}
}
