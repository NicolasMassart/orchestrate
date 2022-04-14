package types

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
)

type CreateWebhookEventStreamRequest struct {
	Name    string            `json:"name" validate:"required" example:"my-webhook-stream"`
	Chain   string            `json:"chain,omitempty" validate:"omitempty" example:"mainnet"`
	URL     string            `json:"url" validate:"required,url" example:"https://my-event-steam-endpoint.com"`
	Headers map[string]string `json:"headers,omitempty" validate:"omitempty"`
	Labels  map[string]string `json:"labels,omitempty" validate:"omitempty"`
}

func (r *CreateWebhookEventStreamRequest) ToEntity() *entities.EventStream {
	return &entities.EventStream{
		Name: r.Name,
		Specs: &entities.Webhook{
			URL:     r.URL,
			Headers: r.Headers,
		},
		Channel: entities.EventStreamChannelWebhook,
		Status:  entities.EventStreamStatusLive,
		Labels:  r.Labels,
	}
}

type CreateKafkaEventStreamRequest struct {
	Name   string            `json:"name" validate:"required" example:"my-kafka-stream"`
	Chain  string            `json:"chain,omitempty" validate:"omitempty" example:"mainnet"`
	Topic  string            `json:"topic" validate:"required" example:"my-notification-topic"`
	Labels map[string]string `json:"labels,omitempty" validate:"omitempty"`
}

func (r *CreateKafkaEventStreamRequest) ToEntity() *entities.EventStream {
	return &entities.EventStream{
		Name: r.Name,
		Specs: &entities.Kafka{
			Topic: r.Topic,
		},
		Channel: entities.EventStreamChannelKafka,
		Status:  entities.EventStreamStatusLive,
		Labels:  r.Labels,
	}
}

type UpdateKafkaEventStreamRequest struct {
	Name   string            `json:"name,omitempty" validate:"omitempty" example:"my-kafka-stream"`
	Chain  string            `json:"chain,omitempty" validate:"omitempty" example:"mainnet"`
	Topic  string            `json:"topic,omitempty" validate:"omitempty" example:"my-notification-topic"`
	Status string            `json:"status,omitempty" validate:"omitempty,isEventStreamStatus" example:"PAUSED"`
	Labels map[string]string `json:"labels,omitempty" validate:"omitempty"`
}

func (r *UpdateKafkaEventStreamRequest) ToEntity(uuid string) *entities.EventStream {
	es := &entities.EventStream{
		UUID:    uuid,
		Name:    r.Name,
		Status:  entities.EventStreamStatus(r.Status),
		Labels:  r.Labels,
		Channel: entities.EventStreamChannelKafka,
	}

	if r.Topic != "" {
		es.Specs = &entities.Kafka{
			Topic: r.Topic,
		}
	}

	return es
}

type UpdateWebhookEventStreamRequest struct {
	Name    string            `json:"name,omitempty" validate:"omitempty" example:"my-kafka-stream"`
	Chain   string            `json:"chain,omitempty" validate:"omitempty" example:"mainnet"`
	URL     string            `json:"url,omitempty" validate:"omitempty,url" example:"https://my-event-steam-endpoint.com"`
	Headers map[string]string `json:"headers,omitempty" validate:"omitempty"`
	Status  string            `json:"status,omitempty" validate:"omitempty,isEventStreamStatus" example:"PAUSED"`
	Labels  map[string]string `json:"labels,omitempty" validate:"omitempty"`
}

func (r *UpdateWebhookEventStreamRequest) ToEntity(uuid string) *entities.EventStream {
	es := &entities.EventStream{
		UUID:    uuid,
		Name:    r.Name,
		Status:  entities.EventStreamStatus(r.Status),
		Labels:  r.Labels,
		Channel: entities.EventStreamChannelWebhook,
	}

	if r.URL != "" || r.Headers != nil {
		es.Specs = &entities.Webhook{
			URL:     r.URL,
			Headers: r.Headers,
		}
	}

	return es
}

type EventStreamResponse struct {
	UUID      string            `json:"uuid"  example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	Name      string            `json:"name"  example:"my-stream"`
	ChainUUID string            `json:"chainUUID,omitempty"  example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	TenantID  string            `json:"tenantID"  example:"foo"`
	OwnerID   string            `json:"ownerID,omitempty"  example:"foo"`
	Specs     interface{}       `json:"specs"`
	Channel   string            `json:"channel"  example:"webhook"`
	Status    string            `json:"status"  example:"LIVE"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"createdAt"  example:"2020-07-09T12:35:42.115395Z"`
	UpdatedAt time.Time         `json:"updatedAt"  example:"2020-07-09T12:35:42.115395Z"`
}

func NewEventStreamResponse(e *entities.EventStream) *EventStreamResponse {
	return &EventStreamResponse{
		UUID:      e.UUID,
		Name:      e.Name,
		ChainUUID: e.ChainUUID,
		TenantID:  e.TenantID,
		OwnerID:   e.OwnerID,
		Specs:     e.Specs,
		Channel:   string(e.Channel),
		Status:    string(e.Status),
		Labels:    e.Labels,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func NewEventStreamResponses(eventStreams []*entities.EventStream) []*EventStreamResponse {
	response := []*EventStreamResponse{}
	for _, e := range eventStreams {
		response = append(response, NewEventStreamResponse(e))
	}

	return response
}
