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
	Specs     interface{}
	Channel   string
	Status    string
	Labels    map[string]string
	ChainUUID string    `pg:"alias:chain_uuid"`
	TenantID  string    `pg:"alias:tenant_id"`
	OwnerID   string    `pg:"alias:owner_id"`
	CreatedAt time.Time `pg:"default:now()"`
	UpdatedAt time.Time `pg:"default:now()"`
}

func NewEventStream(eventStream *entities.EventStream) *EventStream {
	return &EventStream{
		UUID:      eventStream.UUID,
		Name:      eventStream.Name,
		Specs:     eventStream.Specs,
		Channel:   string(eventStream.Channel),
		Status:    string(eventStream.Status),
		ChainUUID: eventStream.ChainUUID,
		Labels:    eventStream.Labels,
		TenantID:  eventStream.TenantID,
		OwnerID:   eventStream.OwnerID,
		CreatedAt: eventStream.CreatedAt,
		UpdatedAt: eventStream.UpdatedAt,
	}
}

func NewEventStreams(eventStreams []*EventStream) []*entities.EventStream {
	res := []*entities.EventStream{}
	for _, e := range eventStreams {
		res = append(res, e.ToEntity())
	}

	return res
}

func (e *EventStream) ToEntity() *entities.EventStream {
	return &entities.EventStream{
		UUID:      e.UUID,
		Name:      e.Name,
		ChainUUID: e.ChainUUID,
		TenantID:  e.TenantID,
		OwnerID:   e.OwnerID,
		Channel:   entities.EventStreamChannel(e.Channel),
		Status:    entities.EventStreamStatus(e.Status),
		Specs:     e.Specs,
		Labels:    e.Labels,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
