package entities

import (
	"time"
)

type EventStreamChannel string
type EventStreamStatus string

const (
	EventStreamChannelWebhook EventStreamChannel = "webhook"
	EventStreamChannelKafka   EventStreamChannel = "kafka"
)

const (
	EventStreamStatusLive    EventStreamStatus = "LIVE"
	EventStreamStatusSuspend EventStreamStatus = "SUSPENDED"
)

const WildcardChainUUID = "00000000-0000-0000-0000-000000000000"
const WildcardChainName = "*"

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
