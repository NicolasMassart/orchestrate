package models

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
)

type EventStream struct {
	tableName struct{} `pg:"event_streams"` // nolint:unused,structcheck // reason

	ID        int `pg:"alias:id"`
	UUID      string
	Name      string
	Specs     *EventStreamSpecs
	Channel   string
	Status    string
	Labels    map[string]string
	ChainUUID string    `pg:"alias:chain_uuid"`
	TenantID  string    `pg:"alias:tenant_id"`
	OwnerID   string    `pg:"alias:owner_id"`
	CreatedAt time.Time `pg:"default:now()"`
	UpdatedAt time.Time `pg:"default:now()"`
}

type EventStreamSpecs struct {
	WebHook *entities.EventStreamWebhookSpec `json:"webhook,omitempty"`
	Kafka   *entities.EventStreamKafkaSpec   `json:"kafka,omitempty"`
}

func NewEventStream(eventStream *entities.EventStream) *EventStream {
	es := &EventStream{
		UUID:      eventStream.UUID,
		Name:      eventStream.Name,
		Channel:   string(eventStream.Channel),
		Status:    string(eventStream.Status),
		ChainUUID: eventStream.ChainUUID,
		Labels:    eventStream.Labels,
		TenantID:  eventStream.TenantID,
		OwnerID:   eventStream.OwnerID,
		CreatedAt: eventStream.CreatedAt,
		UpdatedAt: eventStream.UpdatedAt,
	}

	switch eventStream.Channel {
	case entities.EventStreamChannelKafka:
		if eventStream.Kafka != nil {
			es.Specs = &EventStreamSpecs{
				Kafka: eventStream.Kafka,
			}
		}
	case entities.EventStreamChannelWebhook:
		if eventStream.Webhook != nil {
			es.Specs = &EventStreamSpecs{
				WebHook: eventStream.Webhook,
			}
		}
	}

	return es
}

func NewEventStreams(eventStreams []*EventStream) []*entities.EventStream {
	res := []*entities.EventStream{}
	for _, e := range eventStreams {
		res = append(res, e.ToEntity())
	}

	return res
}

func (e *EventStream) ToEntity() *entities.EventStream {
	es := &entities.EventStream{
		UUID:      e.UUID,
		Name:      e.Name,
		ChainUUID: e.ChainUUID,
		TenantID:  e.TenantID,
		OwnerID:   e.OwnerID,
		Channel:   entities.EventStreamChannel(e.Channel),
		Status:    entities.EventStreamStatus(e.Status),
		Labels:    e.Labels,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	if e.Specs != nil {
		switch es.Channel {
		case entities.EventStreamChannelWebhook:
			es.Webhook = e.Specs.WebHook
		case entities.EventStreamChannelKafka:
			es.Kafka = e.Specs.Kafka
		}
	}

	return es
}
