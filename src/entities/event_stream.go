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
	EventStreamStatusLive    EventStreamStatus = "LIVE"
	EventStreamStatusSuspend EventStreamStatus = "SUSPEND"
)

type EventStream struct {
	UUID      string
	Name      string
	ChainUUID string
	TenantID  string
	OwnerID   string
	Webhook   *EventStreamWebhookSpec
	Kafka     *EventStreamKafkaSpec
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
