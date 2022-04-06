package testdata

import (
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/gofrs/uuid"
)

func fakeEventStream() *entities.EventStream {
	return &entities.EventStream{
		UUID:      uuid.Must(uuid.NewV4()).String(),
		Name:      "my_event_stream",
		ChainUUID: uuid.Must(uuid.NewV4()).String(),
		TenantID:  multitenancy.DefaultTenant,
		OwnerID:   "",
		Status:    entities.EventStreamStatusLive,
	}
}

func FakeWebhookEventStream() *entities.EventStream {
	eventStrean := fakeEventStream()
	eventStrean.Channel = entities.EventStreamChannelWebhook
	eventStrean.Specs = &entities.Webhook{
		URL: "https://mywebhook/1",
		Headers: map[string]string{
			"Authorization": "Bearer jwt",
		},
	}

	return eventStrean
}

func FakeKafkaEventStream() *entities.EventStream {
	eventStrean := fakeEventStream()
	eventStrean.Channel = entities.EventStreamChannelWebhook
	eventStrean.Specs = &entities.Kafka{
		Topic: "my-topic",
	}

	return eventStrean
}
