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

type Webhook struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

type Kafka struct {
	Topic string `json:"topic"`
}

func (e *EventStream) WebHook() *Webhook {
	return e.Specs.(*Webhook) // No need to verify the casting (assertion, not an exception)
}

func (e *EventStream) Kafka() *Kafka {
	return e.Specs.(*Kafka) // No need to verify the casting (assertion, not an exception)
}
