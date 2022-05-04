package testdata

import (
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/quorum-key-manager/pkg/common"
)

func FakeCreateWebhookEventStreamRequest() *api.CreateEventStreamRequest {
	return &api.CreateEventStreamRequest{
		Channel: string(entities.EventStreamChannelWebhook),
		Name:    "eventstream-webhook-" + common.RandString(5),
		Chain:   "mainnet",
		Webhook: &api.WebhookRequest{
			URL: "https://my-webhook-endpoint.com",
			Headers: map[string]string{
				"Authorization": "Bearer JWT",
			},
		},
		Labels: map[string]string{
			"label": "labelValue",
		},
	}
}

func FakeCreateKafkaEventStreamRequest() *api.CreateEventStreamRequest {
	return &api.CreateEventStreamRequest{
		Channel: string(entities.EventStreamChannelKafka),
		Name:    "eventstream-webhook-" + common.RandString(5),
		Chain:   "mainnet",
		Kafka: &api.KafkaRequest{
			Topic: "topic-tx-decoded",
		},
		Labels: map[string]string{
			"label": "labelValue",
		},
	}
}

func FakeUpdateEventStreamRequest() *api.UpdateEventStreamRequest {
	return &api.UpdateEventStreamRequest{
		Status: string(entities.EventStreamStatusSuspend),
	}
}
