package entities

import (
	"time"
)

type EventStreamChannel string
type EventStreamStatus string

var (
	EventStreamChannelWebhook EventStreamChannel = "webhook"
	EventStreamChannelKafka   EventStreamChannel = "kafka"
)

var (
	EventStreamStatusLive   EventStreamStatus = "LIVE"
	EventStreamStatusPaused EventStreamStatus = "PAUSED"
)

type EventStream struct {
	UUID      string
	Name      string
	ChainUUID string
	TenantID  string
	OwnerID   string
	Specs     interface{}
	Channel   EventStreamChannel
	Status    EventStreamStatus
	Labels    map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EventStreamWebhookSpec struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

type EventStreamKafkaSpec struct {
	Topic string `json:"topic"`
}

func (e *EventStream) WebHook() *EventStreamWebhookSpec {
	return e.Specs.(*EventStreamWebhookSpec) // No need to verify the casting (assertion, not an exception)
}

func (e *EventStream) Kafka() *EventStreamKafkaSpec {
	return e.Specs.(*EventStreamKafkaSpec) // No need to verify the casting (assertion, not an exception)
}
